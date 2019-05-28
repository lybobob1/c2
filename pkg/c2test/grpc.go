package c2test

import (
	"errors"
	"fmt"

	e4 "gitlab.com/teserakt/e4common"
	e4test "gitlab.com/teserakt/test-common"
)

func TestGRPCApi(errc chan *e4test.TestResult, grpcClient e4.C2Client) {

	const TESTCLIENTCOUNT = 4
	const TESTTOPICCOUNT = 4
	var testclients [TESTCLIENTCOUNT]e4test.TestClient
	var testtopics [TESTTOPICCOUNT]e4test.TestTopicKey
	var err error

	for i := 0; i < TESTCLIENTCOUNT; i++ {
		err = testclients[i].New()
		if err != nil {
			errc <- &e4test.TestResult{
				Name:     "",
				Result:   false,
				Critical: true,
				Error:    fmt.Errorf("e4test.GenerateID failed. %s", err),
			}
			return
		}
	}
	for i := 0; i < TESTTOPICCOUNT; i++ {
		// we don't actually need keys for these tests;
		// so don't generate them for the topics.
		err = testtopics[i].New(false)
		if err != nil {
			errc <- &e4test.TestResult{
				Name:     "",
				Result:   false,
				Critical: true,
				Error:    fmt.Errorf("e4test.GenerateTopic failed. %s", err),
			}
			return
		}
	}

	for i := 0; i < TESTCLIENTCOUNT; i++ {
		result, err := e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_NEW_CLIENT,
			testclients[i].ID, testclients[i].Name, testclients[i].Key, "", "", 0, 0)
		bresult, ok := result.(bool)
		// must check bresult last, it won't be boolean unless the type assertion
		// succeeds.
		if err != nil || !ok || !bresult {
			errc <- &e4test.TestResult{
				Name:     "CreateClient",
				Result:   false,
				Critical: true,
				Error:    err,
			}
			return
		}
	}
	errc <- &e4test.TestResult{Name: "Add Clients", Result: true, Critical: false, Error: nil}

	for i := 0; i < TESTTOPICCOUNT; i++ {
		result, err := e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_NEW_TOPIC,
			nil, "", nil, testtopics[i].TopicName, "", 0, 0)
		bresult, ok := result.(bool)
		// must check bresult last, it won't be boolean unless the type assertion
		// succeeds.
		if err != nil || !ok || !bresult {
			if err == nil {
				err = errors.New("Type mismatch")
			}
			errc <- &e4test.TestResult{
				Name:     "CreateClient",
				Result:   false,
				Critical: true,
				Error:    err,
			}
			return
		}
	}
	errc <- &e4test.TestResult{Name: "Add Topics", Result: true, Critical: false, Error: nil}

	// *** Add the topic to the client.
	result, err := e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_NEW_TOPIC_CLIENT,
		nil, testclients[0].Name, nil, testtopics[0].TopicName, "", 0, 0)
	bresult, ok := result.(bool)
	if err != nil || !ok || !bresult {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Add Topic to Client",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}
	errc <- &e4test.TestResult{Name: "Add Topic to Client", Result: true, Critical: false, Error: nil}

	// *** Check the M2M link returns the topic we added
	result, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_GET_CLIENT_TOPICS,
		nil, testclients[0].Name, nil, "", "", 0, 10)
	clientTopics, ok := result.([]string)
	if err != nil || !ok {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "M2M Find Added Topic",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}
	if len(clientTopics) != 1 || clientTopics[0] != testtopics[0].TopicName {
		errc <- &e4test.TestResult{
			Name:     "M2M Find Added Topic",
			Result:   false,
			Critical: true,
			Error:    fmt.Errorf("Test M2M Find Added Topic: Incorrect topic returned, returned body is %s", clientTopics),
		}
		return
	}

	errc <- &e4test.TestResult{Name: "M2M Find Added Topic", Result: true, Critical: false, Error: nil}

	// *** Remove the topic from the client (but not the C2)
	result, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_REMOVE_TOPIC_CLIENT,
		nil, testclients[0].Name, nil, testtopics[0].TopicName, "", 0, 10)
	bresult, ok = result.(bool)
	if err != nil || !ok || !bresult {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Remove Topic from Client",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}

	errc <- &e4test.TestResult{Name: "Remove Topic from Client", Result: true, Critical: false, Error: nil}

	// *** Check Topic appears to have been removed from the client
	result, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_GET_CLIENT_TOPICS,
		nil, testclients[0].Name, nil, "", "", 0, 10)
	clientTopics, ok = result.([]string)
	if err != nil || !ok {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Test M2M Doesn't Show Removed Topic",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}
	if len(clientTopics) != 0 {
		errc <- &e4test.TestResult{
			Name:     "Test M2M Doesn't Show Removed Topic",
			Result:   false,
			Critical: true,
			Error:    fmt.Errorf("Test M2M Doesn't Show Removed Topic: Topics found, returned body is %s", clientTopics),
		}
		return
	}
	errc <- &e4test.TestResult{Name: "Test M2M Doesn't Show Removed Topic", Result: true, Critical: false, Error: nil}

	// *** Delete topic
	result, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_REMOVE_TOPIC,
		nil, "", nil, testtopics[0].TopicName, "", 0, 10)
	bresult, ok = result.(bool)
	if err != nil || !ok || !bresult {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Remove topic from C2",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}

	errc <- &e4test.TestResult{Name: "Remove topic from C2", Result: true, Critical: false, Error: nil}

	// *** Check double remove of topic fails
	_, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_REMOVE_TOPIC,
		nil, "", nil, testtopics[0].TopicName, "", 0, 10)
	//bresult, ok = result.(bool)
	if err == nil {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Check double remove fails",
			Result:   false,
			Critical: true,
			Error:    fmt.Errorf("Double remove should report an error via the API and did not"),
		}
		return
	}

	errc <- &e4test.TestResult{Name: "Check double remove fails", Result: true, Critical: false, Error: nil}

	// *** Get topics list
	result, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_GET_TOPICS,
		nil, testclients[0].Name, nil, "", "", 0, 10)
	clientTopics, ok = result.([]string)
	if err != nil || !ok {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Test Fetch Topics",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}
	if len(clientTopics) == 0 || len(clientTopics) != TESTTOPICCOUNT-1 {
		errc <- &e4test.TestResult{
			Name:     "Test Fetch Topics",
			Result:   false,
			Critical: true,
			Error:    fmt.Errorf("Test Fetch Topics: Incorrect number of returned topics, returned body is %s", clientTopics),
		}
		return
	}
	for i := 1; i < TESTTOPICCOUNT; i++ {
		found := false
		testtopic := testtopics[i]
		for j := 0; j < len(clientTopics); j++ {
			if clientTopics[j] == testtopic.TopicName {
				found = true
				break
			}
		}
		if !found {
			errc <- &e4test.TestResult{
				Name:     "Test Fetch Topics",
				Result:   false,
				Critical: true,
				Error:    fmt.Errorf("Test Fetch Topics: Created topic %s not found, topics are %s", testtopic, clientTopics),
			}
			return
		}
	}
	errc <- &e4test.TestResult{Name: "Test Fetch Topics", Result: true, Critical: false, Error: nil}

	// *** Get client list
	result, err = e4test.GrpcC2SendCommand(grpcClient, e4.C2Request_GET_CLIENTS,
		nil, "", nil, "", "", 0, 10)
	clientClients, ok := result.([]string)
	if err != nil || !ok {
		if err == nil {
			err = errors.New("Type mismatch")
		}
		errc <- &e4test.TestResult{
			Name:     "Test Fetch Clients",
			Result:   false,
			Critical: true,
			Error:    err,
		}
		return
	}
	if len(clientClients) == 0 || len(clientClients) != TESTCLIENTCOUNT {
		errc <- &e4test.TestResult{
			Name:     "Test Fetch Clients",
			Result:   false,
			Critical: true,
			Error:    fmt.Errorf("Test Fetch Clients: Incorrect number of returned clients, returned body is %s", clientClients),
		}
		return
	}
	for i := 0; i < TESTCLIENTCOUNT; i++ {
		found := false
		testclient := testclients[i]
		for j := 0; j < len(clientClients); j++ {
			if clientClients[j] == testclient.GetName() {
				found = true
				break
			}
		}
		if !found {
			errc <- &e4test.TestResult{
				Name:     "Test Fetch Client",
				Result:   false,
				Critical: true,
				Error:    fmt.Errorf("Test Fetch Client: Client s%s not found, clients are %s", testclient, clientClients),
			}
			return
		}
	}

	errc <- &e4test.TestResult{Name: "Test Fetch Client", Result: true, Critical: false, Error: nil}

	close(errc)
}
