// Copyright 2020 Teserakt AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/teserakt-io/c2/internal/config"

	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	e4crypto "github.com/teserakt-io/e4go/crypto"

	"github.com/teserakt-io/c2/internal/commands"
	"github.com/teserakt-io/c2/internal/crypto"
	"github.com/teserakt-io/c2/internal/events"
	"github.com/teserakt-io/c2/internal/models"
	"github.com/teserakt-io/c2/internal/protocols"
)

func encryptKey(t *testing.T, dbEncKey []byte, key []byte) []byte {
	protectedkey, err := e4crypto.Encrypt(dbEncKey, nil, key)
	if err != nil {
		t.Fatalf("Failed to encrypt key %v: %v", key, err)
	}

	return protectedkey
}

func newKey(t *testing.T) []byte {
	key := make([]byte, e4crypto.KeyLen)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	return key
}

func createTestClient(t *testing.T, dbEncKey []byte) (models.Client, []byte) {
	clearKey := newKey(t)
	encryptedKey := encryptKey(t, dbEncKey, clearKey)

	randombytes := make([]byte, e4crypto.IDLen)

	_, err := rand.Read(randombytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}

	name := hex.EncodeToString(randombytes)
	id := e4crypto.HashIDAlias(name)

	client := models.Client{
		Name: name,
		E4ID: id,
		Key:  encryptedKey,
	}

	return client, clearKey
}

func createTestTopicKey(t *testing.T, dbEncKey []byte) (models.TopicKey, []byte) {
	clearTopicKey := newKey(t)
	encryptedTopicKey := encryptKey(t, dbEncKey, clearTopicKey)

	topicKey := models.TopicKey{
		Topic: fmt.Sprintf("topic-%d", rand.Int()),
		Key:   encryptedTopicKey,
	}

	return topicKey, clearTopicKey
}

func TestE4(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := models.NewMockDatabase(mockCtrl)
	mockPubSubClient := protocols.NewMockPubSubClient(mockCtrl)
	mockCommandFactory := commands.NewMockFactory(mockCtrl)
	mockEventFactory := events.NewMockFactory(mockCtrl)
	mockEventDispatcher := events.NewMockDispatcher(mockCtrl)
	mockE4Key := crypto.NewMockE4Key(mockCtrl)

	logger := log.New()
	logger.SetOutput(ioutil.Discard)

	dbEncKey := newKey(t)

	cfg := config.CryptoCfg{
		NewClientKeySendPubkey: true,
	}

	service := NewE4(mockDB, mockPubSubClient, mockCommandFactory, mockEventDispatcher, mockEventFactory, mockE4Key, logger, dbEncKey, cfg)
	t.Run("Validation works successfully", func(t *testing.T) {
		names := []string{"test1", "testtest2", "e4test3", "test4", "test5"}

		// test names return the correct hashes:
		for _, name := range names {
			id, err := ValidateE4NameOrIDPair(name, nil)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if bytes.Equal(id, e4crypto.HashIDAlias(name)) == false {
				t.Errorf("Did not return correctly hashed name")
			}
		}

		for _, name := range names {
			submittedID := e4crypto.HashIDAlias(name)
			id, err := ValidateE4NameOrIDPair(name, submittedID)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if bytes.Equal(id, submittedID) == false {
				t.Errorf("Did not return correctly hashed name")
			}
		}

		for _, name := range names {
			submittedID := e4crypto.HashIDAlias(name)
			submittedID[0] ^= 0x01
			_, err := ValidateE4NameOrIDPair(name, submittedID)
			if err == nil {
				t.Errorf("Expected an error, received a non-error result")
			}
			submittedID = e4crypto.HashIDAlias(name)
			shorterID := submittedID[0 : e4crypto.IDLen-2]
			_, err = ValidateE4NameOrIDPair(name, shorterID)
			if err == nil {
				t.Errorf("Expected an error, received a non-error result")
			}
		}

	})

	t.Run("NewClient encrypt key and save properly with name only", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, clearKey := createTestClient(t, dbEncKey)

		mockE4Key.EXPECT().ValidateKey(clearKey).Return(nil).Times(2)
		mockDB.EXPECT().InsertClient(client.Name, client.E4ID, client.Key).Times(2)

		if err := service.NewClient(ctx, client.Name, nil, clearKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if err := service.NewClient(ctx, client.Name, client.E4ID, clearKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RemoveClient deletes the client", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, _ := createTestClient(t, dbEncKey)

		mockDB.EXPECT().DeleteClientByID(client.E4ID)
		if err := service.RemoveClient(ctx, client.E4ID); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewTopicClient links a client to a topic and notify it before updating DB", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, clearClientKey := createTestClient(t, dbEncKey)
		topicKey, clearTopicKey := createTestTopicKey(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		commandPayload := []byte("command-payload")

		expectedEvt := events.Event{
			Type:      events.ClientSubscribed,
			Source:    client.Name,
			Target:    topicKey.Topic,
			Timestamp: time.Now(),
		}

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client.E4ID).Return(client, nil),
			mockDB.EXPECT().GetTopicKey(topicKey.Topic).Return(topicKey, nil),

			mockCommandFactory.EXPECT().CreateSetTopicKeyCommand(topicKey.Topic, clearTopicKey).Return(mockCommand, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearClientKey).Return(commandPayload, nil),

			mockPubSubClient.EXPECT().Publish(gomock.Any(), commandPayload, client, protocols.QoSExactlyOnce),

			mockDB.EXPECT().LinkClientTopic(client, topicKey),

			mockEventFactory.EXPECT().NewClientSubscribedEvent(client.Name, topicKey.Topic).Return(expectedEvt),
			mockEventDispatcher.EXPECT().Dispatch(expectedEvt),
		)

		if err := service.NewTopicClient(ctx, client.E4ID, topicKey.Topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RemoveTopicClient unlink client from topic and notify it before updating DB", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, clearIDKey := createTestClient(t, dbEncKey)
		topicKey, _ := createTestTopicKey(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		commandPayload := []byte("command-payload")

		expectedEvt := events.Event{
			Type:      events.ClientSubscribed,
			Source:    client.Name,
			Target:    topicKey.Topic,
			Timestamp: time.Now(),
		}

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client.E4ID).Return(client, nil),
			mockDB.EXPECT().GetTopicKey(topicKey.Topic).Return(topicKey, nil),

			mockCommandFactory.EXPECT().CreateRemoveTopicCommand(topicKey.Topic).Return(mockCommand, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearIDKey).Return(commandPayload, nil),

			mockPubSubClient.EXPECT().Publish(gomock.Any(), commandPayload, client, protocols.QoSExactlyOnce),

			mockDB.EXPECT().UnlinkClientTopic(client, topicKey),

			mockEventFactory.EXPECT().NewClientUnsubscribedEvent(client.Name, topicKey.Topic).Return(expectedEvt),
			mockEventDispatcher.EXPECT().Dispatch(expectedEvt),
		)

		if err := service.RemoveTopicClient(ctx, client.E4ID, topicKey.Topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ResetClient send a reset command to client", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, clearIDKey := createTestClient(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		commandPayload := []byte("command-payload")

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client.E4ID).Return(client, nil),

			mockCommandFactory.EXPECT().CreateResetTopicsCommand().Return(mockCommand, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearIDKey).Return(commandPayload, nil),

			mockPubSubClient.EXPECT().Publish(gomock.Any(), commandPayload, client, protocols.QoSExactlyOnce),
		)

		if err := service.ResetClient(ctx, client.E4ID); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewTopic creates a new topic, and enable its monitoring", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		topic := "topic"

		mockCommand := commands.NewMockCommand(mockCtrl)
		mockCommand.EXPECT().Type().AnyTimes().Return(byte(1), nil)

		mockTx := models.NewMockDatabase(mockCtrl)

		gomock.InOrder(
			mockPubSubClient.EXPECT().ValidateTopic(topic).Return(nil),
			mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil),
			mockTx.EXPECT().InsertTopicKey(topic, gomock.Any()),
			mockCommandFactory.EXPECT().CreateSetTopicKeyCommand(topic, gomock.Any()).Return(mockCommand, nil),
			mockTx.EXPECT().CountClientsForTopic(topic).Return(0, nil),
			mockTx.EXPECT().CommitTx(),

			mockPubSubClient.EXPECT().SubscribeToTopic(gomock.Any(), topic),
		)

		if err := service.NewTopic(ctx, topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewTopic send the topic key to its clients in batch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		NewTopicBatchSize = 3
		totalClients := 5

		client11, client11Key := createTestClient(t, dbEncKey)
		client11Payload := []byte("client11")
		client12, client12Key := createTestClient(t, dbEncKey)
		client12Payload := []byte("client12")
		client13, client13Key := createTestClient(t, dbEncKey)
		client13Payload := []byte("client13")
		clientsBatch1 := []models.Client{client11, client12, client13}

		client21, client21Key := createTestClient(t, dbEncKey)
		client21Payload := []byte("client21")
		client22, client22Key := createTestClient(t, dbEncKey)
		client22Payload := []byte("client22")
		clientsBatch2 := []models.Client{client21, client22}

		topic := "topic"

		mockCommand := commands.NewMockCommand(mockCtrl)
		mockCommand.EXPECT().Type().AnyTimes().Return(byte(1), nil)

		mockTx := models.NewMockDatabase(mockCtrl)
		gomock.InOrder(
			mockPubSubClient.EXPECT().ValidateTopic(topic).Return(nil),
			mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil),
			mockTx.EXPECT().InsertTopicKey(topic, gomock.Any()),
			mockCommandFactory.EXPECT().CreateSetTopicKeyCommand(topic, gomock.Any()).Return(mockCommand, nil),
			mockTx.EXPECT().CountClientsForTopic(topic).Return(totalClients, nil),
			mockTx.EXPECT().CommitTx(),

			mockDB.EXPECT().GetClientsForTopic(topic, 0, NewTopicBatchSize).Return(clientsBatch1, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client11Key).Return(client11Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client11Payload, client11, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client12Key).Return(client12Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client12Payload, client12, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client13Key).Return(client13Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client13Payload, client13, protocols.QoSExactlyOnce),
			mockDB.EXPECT().GetClientsForTopic(topic, NewTopicBatchSize, NewTopicBatchSize).Return(clientsBatch2, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client21Key).Return(client21Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client21Payload, client21, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client22Key).Return(client22Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client22Payload, client22, protocols.QoSExactlyOnce),
			mockPubSubClient.EXPECT().SubscribeToTopic(gomock.Any(), topic),
		)

		if err := service.NewTopic(ctx, topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewTopic send the topic key to its clients in batch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		NewTopicBatchSize = 3
		totalClients := 5

		client11, client11Key := createTestClient(t, dbEncKey)
		client11Payload := []byte("client11")
		client12, client12Key := createTestClient(t, dbEncKey)
		client12Payload := []byte("client12")
		client13, client13Key := createTestClient(t, dbEncKey)
		client13Payload := []byte("client13")
		clientsBatch1 := []models.Client{client11, client12, client13}

		client21, client21Key := createTestClient(t, dbEncKey)
		client21Payload := []byte("client21")
		client22, client22Key := createTestClient(t, dbEncKey)
		client22Payload := []byte("client22")
		clientsBatch2 := []models.Client{client21, client22}

		topic := "topic"

		mockCommand := commands.NewMockCommand(mockCtrl)
		mockCommand.EXPECT().Type().AnyTimes().Return(byte(1), nil)

		mockTx := models.NewMockDatabase(mockCtrl)
		gomock.InOrder(
			mockPubSubClient.EXPECT().ValidateTopic(topic).Return(nil),
			mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil),
			mockTx.EXPECT().InsertTopicKey(topic, gomock.Any()),
			mockCommandFactory.EXPECT().CreateSetTopicKeyCommand(topic, gomock.Any()).Return(mockCommand, nil),
			mockTx.EXPECT().CountClientsForTopic(topic).Return(totalClients, nil),
			mockTx.EXPECT().CommitTx(),

			mockDB.EXPECT().GetClientsForTopic(topic, 0, NewTopicBatchSize).Return(clientsBatch1, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client11Key).Return(client11Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client11Payload, client11, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client12Key).Return(client12Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client12Payload, client12, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client13Key).Return(client13Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client13Payload, client13, protocols.QoSExactlyOnce),
			mockDB.EXPECT().GetClientsForTopic(topic, NewTopicBatchSize, NewTopicBatchSize).Return(clientsBatch2, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client21Key).Return(client21Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client21Payload, client21, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client22Key).Return(client22Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client22Payload, client22, protocols.QoSExactlyOnce),
			mockPubSubClient.EXPECT().SubscribeToTopic(gomock.Any(), topic),
		)

		if err := service.NewTopic(ctx, topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewTopic send the topic key to its clients in batch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		NewTopicBatchSize = 3
		totalClients := 5

		client11, client11Key := createTestClient(t, dbEncKey)
		client11Payload := []byte("client11")
		client12, client12Key := createTestClient(t, dbEncKey)
		client12Payload := []byte("client12")
		client13, client13Key := createTestClient(t, dbEncKey)
		client13Payload := []byte("client13")
		clientsBatch1 := []models.Client{client11, client12, client13}

		client21, client21Key := createTestClient(t, dbEncKey)
		client21Payload := []byte("client21")
		client22, client22Key := createTestClient(t, dbEncKey)
		client22Payload := []byte("client22")
		clientsBatch2 := []models.Client{client21, client22}

		topic := "topic"

		mockCommand := commands.NewMockCommand(mockCtrl)
		mockCommand.EXPECT().Type().AnyTimes().Return(byte(1), nil)

		mockTx := models.NewMockDatabase(mockCtrl)
		gomock.InOrder(
			mockPubSubClient.EXPECT().ValidateTopic(topic).Return(nil),
			mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil),
			mockTx.EXPECT().InsertTopicKey(topic, gomock.Any()),
			mockCommandFactory.EXPECT().CreateSetTopicKeyCommand(topic, gomock.Any()).Return(mockCommand, nil),
			mockTx.EXPECT().CountClientsForTopic(topic).Return(totalClients, nil),
			mockTx.EXPECT().CommitTx(),

			mockDB.EXPECT().GetClientsForTopic(topic, 0, NewTopicBatchSize).Return(clientsBatch1, nil),

			mockE4Key.EXPECT().ProtectCommand(mockCommand, client11Key).Return(client11Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client11Payload, client11, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client12Key).Return(client12Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client12Payload, client12, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client13Key).Return(client13Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client13Payload, client13, protocols.QoSExactlyOnce),
			mockDB.EXPECT().GetClientsForTopic(topic, NewTopicBatchSize, NewTopicBatchSize).Return(clientsBatch2, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client21Key).Return(client21Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client21Payload, client21, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, client22Key).Return(client22Payload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), client22Payload, client22, protocols.QoSExactlyOnce),
			mockPubSubClient.EXPECT().SubscribeToTopic(gomock.Any(), topic),
		)

		if err := service.NewTopic(ctx, topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RemoveTopic cancel topic monitoring and removes it from DB", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		topic := "topic"

		gomock.InOrder(
			mockPubSubClient.EXPECT().UnsubscribeFromTopic(gomock.Any(), topic),
			mockDB.EXPECT().DeleteTopicKey(topic),
		)

		if err := service.RemoveTopic(ctx, topic); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewClientKey generates a new key, send it to the client and update the DB", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, clearClientKey := createTestClient(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		commandPayload := []byte("command-payload")

		clientKey := []byte("clientKey")
		c2StoredKey := []byte("c2StoredKey")
		protectedC2StoredKey, err := e4crypto.Encrypt(dbEncKey, nil, c2StoredKey)
		if err != nil {
			t.Fatalf("failed to encrypt new key: %v", err)
		}

		mockE4Key.EXPECT().IsPubKeyMode().Return(false)

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client.E4ID).Return(client, nil),
			mockE4Key.EXPECT().RandomKey().Return(clientKey, c2StoredKey, nil),
			mockCommandFactory.EXPECT().CreateSetIDKeyCommand(clientKey).Return(mockCommand, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearClientKey).Return(commandPayload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), commandPayload, client, protocols.QoSExactlyOnce),
			mockDB.EXPECT().InsertClient(client.Name, client.E4ID, protectedC2StoredKey),
		)

		if err := service.NewClientKey(ctx, client.E4ID); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NewClientKey send the new pubkey to linked clients in pubkey mode", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, clearClientKey := createTestClient(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		commandPayload := []byte("command-payload")
		mockSetPubKeyCommand := commands.NewMockCommand(mockCtrl)
		setPubKeyCommandPayload1 := []byte("set-pubkey1")
		setPubKeyCommandPayload2 := []byte("set-pubkey2")

		clientKey := []byte("clientKey")
		c2StoredKey := []byte("c2StoredKey")
		protectedC2StoredKey, err := e4crypto.Encrypt(dbEncKey, nil, c2StoredKey)
		if err != nil {
			t.Fatalf("failed to encrypt new key: %v", err)
		}

		linkedClient1, linkedClient1ClearKey := createTestClient(t, dbEncKey)
		linkedClient2, linkedClient2ClearKey := createTestClient(t, dbEncKey)

		linkedClients := []models.Client{linkedClient1, linkedClient2}

		mockE4Key.EXPECT().IsPubKeyMode().Return(true)

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client.E4ID).Return(client, nil),
			mockE4Key.EXPECT().RandomKey().Return(clientKey, c2StoredKey, nil),
			mockCommandFactory.EXPECT().CreateSetIDKeyCommand(clientKey).Return(mockCommand, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearClientKey).Return(commandPayload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), commandPayload, client, protocols.QoSExactlyOnce),
			mockDB.EXPECT().InsertClient(client.Name, client.E4ID, protectedC2StoredKey),
			// Gomock fail to compare []byte arguments so we check it ourselves
			mockCommandFactory.EXPECT().CreateSetPubKeyCommand(gomock.Any(), client.Name).DoAndReturn(func(key []byte, name string) (commands.Command, error) {
				if !bytes.Equal(key, c2StoredKey) {
					t.Fatalf("invalid public key, got %v, want %v", key, c2StoredKey)
				}
				return mockSetPubKeyCommand, nil
			}),
			mockDB.EXPECT().CountLinkedClients(client.E4ID).Return(2, nil),
			mockDB.EXPECT().GetLinkedClientsForClientByID(client.E4ID, 0, GetLinkedClientsBatchSize).Return(linkedClients, nil),
			mockE4Key.EXPECT().ProtectCommand(mockSetPubKeyCommand, linkedClient1ClearKey).Return(setPubKeyCommandPayload1, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), setPubKeyCommandPayload1, linkedClient1, protocols.QoSExactlyOnce),
			mockE4Key.EXPECT().ProtectCommand(mockSetPubKeyCommand, linkedClient2ClearKey).Return(setPubKeyCommandPayload2, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), setPubKeyCommandPayload2, linkedClient2, protocols.QoSExactlyOnce),
		)

		if err := service.NewClientKey(ctx, client.E4ID); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("CountTopicsForClient return topic count", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, _ := createTestClient(t, dbEncKey)

		expectedCount := 10

		mockDB.EXPECT().CountTopicsForClientByID(client.E4ID).Return(expectedCount, nil)

		count, err := service.CountTopicsForClient(ctx, client.E4ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if count != expectedCount {
			t.Errorf("Expected count to be %d, got %d", expectedCount, count)
		}
	})

	t.Run("GetTopicsForClient returns topics for a given ID", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, _ := createTestClient(t, dbEncKey)
		expectedOffset := 1
		expectedCount := 2

		t1, _ := createTestTopicKey(t, dbEncKey)
		t2, _ := createTestTopicKey(t, dbEncKey)
		t3, _ := createTestTopicKey(t, dbEncKey)

		topicKeys := []models.TopicKey{t1, t2, t3}
		expectedTopics := []string{t1.Topic, t2.Topic, t3.Topic}

		mockDB.EXPECT().GetTopicsForClientByID(client.E4ID, expectedOffset, expectedCount).Return(topicKeys, nil)

		topics, err := service.GetTopicsRangeByClient(ctx, client.E4ID, expectedOffset, expectedCount)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if reflect.DeepEqual(topics, expectedTopics) == false {
			t.Errorf("Expected topics to be %v, got %v", expectedTopics, topics)
		}
	})

	t.Run("GetTopicsForClient returns an empty slice when no results", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockDB.EXPECT().GetTopicsForClientByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
		topics, err := service.GetTopicsRangeByClient(ctx, e4crypto.HashIDAlias("client"), 1, 2)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if topics == nil {
			t.Errorf("Expected topics to be an empty slice, got nil")
		}
	})

	t.Run("CountClientsForTopic returns the IDs count for a given topic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		topicKey, _ := createTestTopicKey(t, dbEncKey)

		expectedCount := 10

		mockDB.EXPECT().CountClientsForTopic(topicKey.Topic).Return(expectedCount, nil)

		count, err := service.CountClientsForTopic(ctx, topicKey.Topic)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if count != expectedCount {
			t.Errorf("Expected count to be %d, got %d", expectedCount, count)
		}
	})

	t.Run("GetClientsByTopic returns all clients for a given topic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		topicKey, _ := createTestTopicKey(t, dbEncKey)
		expectedOffset := 1
		expectedCount := 2

		i1, _ := createTestClient(t, dbEncKey)
		i2, _ := createTestClient(t, dbEncKey)
		i3, _ := createTestClient(t, dbEncKey)

		clients := []models.Client{i1, i2, i3}
		expectedIDNamePairs := []IDNamePair{
			IDNamePair{Name: i1.Name, ID: i1.E4ID},
			IDNamePair{Name: i2.Name, ID: i2.E4ID},
			IDNamePair{Name: i3.Name, ID: i3.E4ID},
		}

		mockDB.EXPECT().GetClientsForTopic(topicKey.Topic, expectedOffset, expectedCount).Return(clients, nil)

		idNamePairs, err := service.GetClientsRangeByTopic(ctx, topicKey.Topic, expectedOffset, expectedCount)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if reflect.DeepEqual(idNamePairs, expectedIDNamePairs) == false {
			t.Errorf("Expected idNamePairs to be %v, got %v", expectedIDNamePairs, idNamePairs)
		}
	})

	t.Run("GetClientsByTopic returns an empty slice when no results", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockDB.EXPECT().GetClientsForTopic(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
		idNamePairs, err := service.GetClientsRangeByTopic(ctx, "topic", 1, 2)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if idNamePairs == nil {
			t.Errorf("Expected idNamePairs to be an empty slice, got nil")
		}
	})

	t.Run("GetClientsRange returns client ID and Name pairs rom offset and count", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		i1, _ := createTestClient(t, dbEncKey)
		i2, _ := createTestClient(t, dbEncKey)
		i3, _ := createTestClient(t, dbEncKey)

		clients := []models.Client{i1, i2, i3}
		expectedPäirs := []IDNamePair{
			IDNamePair{Name: i1.Name, ID: i1.E4ID},
			IDNamePair{Name: i2.Name, ID: i2.E4ID},
			IDNamePair{Name: i3.Name, ID: i3.E4ID},
		}

		offset := 1
		count := 2

		mockDB.EXPECT().GetClientsRange(offset, count).Return(clients, nil)

		idNamePairs, err := service.GetClientsRange(ctx, offset, count)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if reflect.DeepEqual(idNamePairs, expectedPäirs) == false {
			t.Errorf("Expected idNamePairs to be %#v, got %#v", expectedPäirs, idNamePairs)
		}
	})

	t.Run("GetClientsRange returns an empty slice when no results", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockDB.EXPECT().GetClientsRange(gomock.Any(), gomock.Any()).Return(nil, nil)
		idNamePairs, err := service.GetClientsRange(ctx, 1, 2)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if idNamePairs == nil {
			t.Errorf("Expected idNamePairs to be an empty slice, got nil")
		}
	})

	t.Run("GetTopicsRange returns topics from offset and count", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		t1, _ := createTestTopicKey(t, dbEncKey)
		t2, _ := createTestTopicKey(t, dbEncKey)
		t3, _ := createTestTopicKey(t, dbEncKey)

		topicKeys := []models.TopicKey{t1, t2, t3}
		expectedTopics := []string{
			t1.Topic,
			t2.Topic,
			t3.Topic,
		}

		offset := 1
		count := 2

		mockDB.EXPECT().GetTopicsRange(offset, count).Return(topicKeys, nil)

		topics, err := service.GetTopicsRange(ctx, offset, count)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if reflect.DeepEqual(topics, expectedTopics) == false {
			t.Errorf("Expected topics to be %#v, got %#v", expectedTopics, topics)
		}
	})

	t.Run("GetTopicsRange returns an empty slice when no results", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockDB.EXPECT().GetTopicsRange(gomock.Any(), gomock.Any()).Return(nil, nil)
		topics, err := service.GetTopicsRange(ctx, 1, 2)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if topics == nil {
			t.Errorf("Expected topics to be an empty slice, got nil")
		}
	})

	t.Run("CountClients returns client count", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		expectedCount := 42
		mockDB.EXPECT().CountClients().Return(expectedCount, nil)

		c, err := service.CountClients(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if c != expectedCount {
			t.Errorf("Expected count to be %d, got %d", expectedCount, c)
		}
	})

	t.Run("CountTopics returns client count", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		expectedCount := 42
		mockDB.EXPECT().CountTopicKeys().Return(expectedCount, nil)

		c, err := service.CountTopics(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if c != expectedCount {
			t.Errorf("Expected count to be %d, got %d", expectedCount, c)
		}
	})

	t.Run("LinkClient properly links clients together", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client1, _ := createTestClient(t, dbEncKey)
		client2, _ := createTestClient(t, dbEncKey)

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client1.E4ID).Return(client1, nil),
			mockDB.EXPECT().GetClientByID(client2.E4ID).Return(client2, nil),
			mockDB.EXPECT().LinkClient(client1, client2).Return(nil),
		)

		if err := service.LinkClient(ctx, client1.E4ID, client2.E4ID); err != nil {
			t.Fatalf("failed to link clients: %v", err)
		}
	})

	t.Run("UnlinkClient properly unlink clients", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client1, _ := createTestClient(t, dbEncKey)
		client2, _ := createTestClient(t, dbEncKey)

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(client1.E4ID).Return(client1, nil),
			mockDB.EXPECT().GetClientByID(client2.E4ID).Return(client2, nil),
			mockDB.EXPECT().UnlinkClient(client1, client2).Return(nil),
		)

		if err := service.UnlinkClient(ctx, client1.E4ID, client2.E4ID); err != nil {
			t.Fatalf("failed to link clients: %v", err)
		}
	})

	t.Run("CountLinkedClients return the expected client count", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client1, _ := createTestClient(t, dbEncKey)

		expectedCount := 42
		mockDB.EXPECT().CountLinkedClients(client1.E4ID).Return(expectedCount, nil)
		count, err := service.CountLinkedClients(ctx, client1.E4ID)
		if err != nil {
			t.Fatalf("failed to link clients: %v", err)
		}
		if count != expectedCount {
			t.Fatalf("invalid count, got %d, want %d", count, expectedCount)
		}
	})

	t.Run("GetLinkedClients returns the linked clients", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client1, _ := createTestClient(t, dbEncKey)
		client2, _ := createTestClient(t, dbEncKey)
		client3, _ := createTestClient(t, dbEncKey)

		expectedPairs := []IDNamePair{
			IDNamePair{ID: client2.E4ID, Name: client2.Name},
			IDNamePair{ID: client3.E4ID, Name: client3.Name},
		}
		expectedOffset := 1
		expectedCount := 2

		mockDB.EXPECT().GetLinkedClientsForClientByID(client1.E4ID, expectedOffset, expectedCount).Return([]models.Client{client2, client3}, nil)

		pairs, err := service.GetLinkedClients(ctx, client1.E4ID, expectedOffset, expectedCount)
		if err != nil {
			t.Fatalf("failed to link clients: %v", err)
		}

		if !reflect.DeepEqual(pairs, expectedPairs) {
			t.Fatalf("invalid linked pairs returned, got %#v, want %#v", pairs, expectedPairs)
		}
	})

	t.Run("SendClientPubKey returns error when key does not support pubkey mode", func(t *testing.T) {
		mockE4Key.EXPECT().IsPubKeyMode().Return(false)
		err := service.SendClientPubKey(context.Background(), []byte("source"), []byte("target"))
		want := ErrInvalidCryptoMode{}
		if err != want {
			t.Fatalf("got error %v, wanted %v", err, want)
		}
	})

	t.Run("SendClientPubKey sends the expected command with a key supporting pubkey mode", func(t *testing.T) {
		mockE4Key.EXPECT().IsPubKeyMode().Return(true)

		sourceClient, clearSourceClientKey := createTestClient(t, dbEncKey)
		targetClient, clearTargetClientKey := createTestClient(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		cmdPayload := []byte("protectedSetPubKeyCommand")

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(sourceClient.E4ID).Return(sourceClient, nil),
			mockDB.EXPECT().GetClientByID(targetClient.E4ID).Return(targetClient, nil),

			// TODO: figure out why does gomock doesn't match properly the key here ? DoAndReturn used as a workaround...
			mockCommandFactory.EXPECT().CreateSetPubKeyCommand(gomock.Any(), sourceClient.Name).DoAndReturn(func(key []byte, clientName string) (commands.Command, error) {
				if !bytes.Equal(clearSourceClientKey, key) {
					t.Fatalf("Invalid key, got %v, want %v", key, clearSourceClientKey)
				}
				return mockCommand, nil
			}),

			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearTargetClientKey).Return(cmdPayload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), cmdPayload, targetClient, protocols.QoSExactlyOnce).Return(nil),
		)

		err := service.SendClientPubKey(context.Background(), sourceClient.E4ID, targetClient.E4ID)
		if err != nil {
			t.Fatalf("failed to send pubkey command: %v", err)
		}
	})

	t.Run("RemoveClientPubKey returns error when key does not support pubkey mode", func(t *testing.T) {
		mockE4Key.EXPECT().IsPubKeyMode().Return(false)
		err := service.RemoveClientPubKey(context.Background(), []byte("source"), []byte("target"))
		want := ErrInvalidCryptoMode{}
		if err != want {
			t.Fatalf("got error %v, wanted %v", err, want)
		}
	})

	t.Run("RemoveClientPubKey sends the expected command to the target client", func(t *testing.T) {
		mockE4Key.EXPECT().IsPubKeyMode().Return(true)

		sourceClient, _ := createTestClient(t, dbEncKey)
		targetClient, clearTargetClientKey := createTestClient(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		cmdPayload := []byte("protectedRemovePubKeyCommand")

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(sourceClient.E4ID).Return(sourceClient, nil),
			mockDB.EXPECT().GetClientByID(targetClient.E4ID).Return(targetClient, nil),

			mockCommandFactory.EXPECT().CreateRemovePubKeyCommand(sourceClient.Name).Return(mockCommand, nil),

			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearTargetClientKey).Return(cmdPayload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), cmdPayload, targetClient, protocols.QoSExactlyOnce).Return(nil),
		)

		err := service.RemoveClientPubKey(context.Background(), sourceClient.E4ID, targetClient.E4ID)
		if err != nil {
			t.Fatalf("failed to send RemovePubKey command: %v", err)
		}
	})

	t.Run("ResetClientPubKeys sends the expected command to the target client", func(t *testing.T) {
		mockE4Key.EXPECT().IsPubKeyMode().Return(true)

		targetClient, clearTargetClientKey := createTestClient(t, dbEncKey)

		mockCommand := commands.NewMockCommand(mockCtrl)
		cmdPayload := []byte("protectedResetClientPubKeysCommand")

		gomock.InOrder(
			mockDB.EXPECT().GetClientByID(targetClient.E4ID).Return(targetClient, nil),

			mockCommandFactory.EXPECT().CreateResetPubKeysCommand().Return(mockCommand, nil),

			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearTargetClientKey).Return(cmdPayload, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), cmdPayload, targetClient, protocols.QoSExactlyOnce).Return(nil),
		)

		err := service.ResetClientPubKeys(context.Background(), targetClient.E4ID)
		if err != nil {
			t.Fatalf("failed to send ResetPubKeys command: %v", err)
		}
	})

	t.Run("NewC2Keys returns error when key does not support pubkey mode", func(t *testing.T) {
		mockE4Key.EXPECT().IsPubKeyMode().Return(false)
		err := service.NewC2Key(context.Background())
		want := ErrInvalidCryptoMode{}
		if err != want {
			t.Fatalf("got error %v, wanted %v", err, want)
		}
	})

	t.Run("NewC2Keys send a new C2 public key to all clients", func(t *testing.T) {
		NewC2KeyBatchSize = 2

		mockE4Key.EXPECT().IsPubKeyMode().Return(true)

		newC2PubKey := []byte("newC2PubKey")

		mockCommand := commands.NewMockCommand(mockCtrl)
		cmdPayload1 := []byte("setC2KeyCommand1")
		cmdPayload2 := []byte("setC2KeyCommand2")
		cmdPayload3 := []byte("setC2KeyCommand3")

		client1, clearClient1Key := createTestClient(t, dbEncKey)
		client2, clearClient2Key := createTestClient(t, dbEncKey)
		client3, clearClient3Key := createTestClient(t, dbEncKey)

		clientsBatch1 := []models.Client{client1, client2}
		clientsBatch2 := []models.Client{client3}

		mockC2Tx := crypto.NewMockC2KeyRotationTx(mockCtrl)
		mockC2Tx.EXPECT().GetNewPublicKey().AnyTimes().Return(newC2PubKey)

		gomock.InOrder(
			mockE4Key.EXPECT().NewC2KeyRotationTx().Return(mockC2Tx, nil),
			mockCommandFactory.EXPECT().CreateSetC2KeyCommand(newC2PubKey).Return(mockCommand, nil),

			mockDB.EXPECT().GetClientsRange(0, NewC2KeyBatchSize).Return(clientsBatch1, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearClient1Key).Return(cmdPayload1, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), cmdPayload1, client1, protocols.QoSExactlyOnce).Return(nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearClient2Key).Return(cmdPayload2, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), cmdPayload2, client2, protocols.QoSExactlyOnce).Return(nil),

			mockDB.EXPECT().GetClientsRange(NewC2KeyBatchSize, NewC2KeyBatchSize).Return(clientsBatch2, nil),
			mockE4Key.EXPECT().ProtectCommand(mockCommand, clearClient3Key).Return(cmdPayload3, nil),
			mockPubSubClient.EXPECT().Publish(gomock.Any(), cmdPayload3, client3, protocols.QoSExactlyOnce).Return(nil),

			mockC2Tx.EXPECT().Commit().Return(nil),
		)

		err := service.NewC2Key(context.Background())
		if err != nil {
			t.Fatalf("failed to set new C2 key: %v", err)
		}
	})

	t.Run("ProtectMessage protect a message with a topic key", func(t *testing.T) {
		topic, clearTopicKey := createTestTopicKey(t, dbEncKey)
		expectedClearData := []byte("clear-data")
		expectedProtectedData, err := e4crypto.ProtectSymKey(expectedClearData, clearTopicKey)
		if err != nil {
			t.Fatalf("failed to protect test payload: %v", err)
		}

		mockDB.EXPECT().GetTopicKey(topic.Topic).Return(topic, nil)

		protectedData, err := service.ProtectMessage(context.Background(), topic.Topic, expectedClearData)
		if err != nil {
			t.Fatalf("failed to protect message: %v", err)
		}

		if !bytes.Equal(protectedData, expectedProtectedData) {
			t.Fatalf("invalid protected data, got %x, want %x", protectedData, expectedProtectedData)
		}
	})

	t.Run("UnprotectMessage unprotect a message with a topic key", func(t *testing.T) {
		topic, clearTopicKey := createTestTopicKey(t, dbEncKey)
		expectedClearData := []byte("clear-data")
		expectedProtectedData, err := e4crypto.ProtectSymKey(expectedClearData, clearTopicKey)
		if err != nil {
			t.Fatalf("failed to protect test payload: %v", err)
		}

		mockDB.EXPECT().GetTopicKey(topic.Topic).Return(topic, nil)

		clearData, err := service.UnprotectMessage(context.Background(), topic.Topic, expectedProtectedData)
		if err != nil {
			t.Fatalf("failed to protect message: %v", err)
		}

		if !bytes.Equal(clearData, expectedClearData) {
			t.Fatalf("invalid clear data, got %x, want %x", clearData, expectedClearData)
		}
	})
}
