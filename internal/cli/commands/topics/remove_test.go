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

package topics

import (
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/teserakt-io/c2/internal/cli"
	"github.com/teserakt-io/c2/pkg/pb"
)

func newTestRemoveCommand(clientFactory cli.APIClientFactory) cli.Command {
	cmd := NewRemoveCommand(clientFactory)
	cmd.CobraCmd().SetOutput(ioutil.Discard)
	cmd.CobraCmd().DisableFlagParsing = true

	return cmd
}

func TestRemove(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	c2Client := cli.NewMockC2Client(mockCtrl)

	c2ClientFactory := cli.NewMockAPIClientFactory(mockCtrl)
	c2ClientFactory.EXPECT().NewClient(gomock.Any()).AnyTimes().Return(c2Client, nil)

	t.Run("Execute properly checks flags and return expected errors", func(t *testing.T) {
		badFlagsDataset := []map[string]string{
			// No name
			map[string]string{},
		}

		for _, flagData := range badFlagsDataset {
			cmd := newTestRemoveCommand(c2ClientFactory)
			for name, value := range flagData {
				cmd.CobraCmd().Flags().Set(name, value)
			}
			err := cmd.CobraCmd().Execute()
			if err == nil {
				t.Error("Expected an error, got nil")
			}
		}
	})

	t.Run("Execute forward expected request to the c2Client when passing a name", func(t *testing.T) {
		expectedTopic := "testTopic1"
		expectedRequest := &pb.RemoveTopicRequest{
			Topic: expectedTopic,
		}

		c2Client.EXPECT().RemoveTopic(gomock.Any(), expectedRequest).Return(&pb.RemoveTopicResponse{}, nil)
		c2Client.EXPECT().Close()

		cmd := newTestRemoveCommand(c2ClientFactory)
		cmd.CobraCmd().Flags().Set("name", expectedTopic)
		err := cmd.CobraCmd().Execute()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}
