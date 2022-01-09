package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"latencychecker/chat/model"
	"os"
	"sync"
	"time"

	"github.com/ably/ably-go/ably"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

const APPNAME = "latencyChecker"
const FILEOUTPUTEXTENSION = "csv"

func Start(apiKey, ch, name, fileLocation string, messageNum, delay, wait int) {
	log = logrus.New()
	mu := &sync.Mutex{}
	results := map[string]map[string][]model.Payload{}

	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stderr)

	log.Formatter.(*logrus.TextFormatter).FullTimestamp = true

	log.Infof("Starting %s for client %s", APPNAME, name)

	client, err := ably.NewRealtime(
		ably.WithKey(apiKey),
		ably.WithClientID(name),
		ably.WithEchoMessages(false),
	)

	if err != nil {
		log.Fatalf("Couldn't create ably client, %v", err)
	}

	channel := client.Channels.Get(ch)

	subscribe(channel, results, mu, name)

	for i := 0; i < messageNum-1; i++ {
		publishing(channel, createMsg(name, i))
		time.Sleep(time.Duration(delay) * time.Second)
	}

	publishing(channel, createMsg(name, messageNum-1))
	time.Sleep(time.Duration(wait) * time.Second)

	if err := channel.Detach(context.Background()); err != nil {
		log.Errorf("Error cleaning ably resources: %v", err)
	}

	mu.Lock()
	results = filterTimedOutMgs(results, wait)
	mu.Unlock()

	fileName := fmt.Sprintf("%s/%s_%s_%d.%s", fileLocation, APPNAME, name, time.Now().Unix(), FILEOUTPUTEXTENSION)
	f, err := os.Create(fileName)

	if err != nil {
		log.Errorf("Couldn't create file at path %s, %v", fileLocation, err)
	}

	defer f.Close()

	log.Infof("Created file %s at location %s", fileName, fileLocation)

	mu.Lock()
	if err := createResultsTable(f, results); err != nil {
		log.Infof("Couldn't write table to file, %v", err)

		if err := os.Remove(fileName); err != nil {
			log.Errorf("Error while deleting file %s %v", fileName, err)
		}

		return
	}
	mu.Unlock()

	if err := client.Channels.Release(context.Background(), ch); err != nil {
		log.Errorf("Error cleaning ably resources: %v", err)
	}

	log.Infof("Exiting latencyChecker for client %s", name)
}

func publishing(channel *ably.RealtimeChannel, msg model.Payload) {
	err := channel.Publish(context.Background(), "message", msg)
	if err != nil {
		log.Errorf("Publishing to channel, %v", err)
	}
}

func subscribe(channel *ably.RealtimeChannel, results map[string]map[string][]model.Payload, mu *sync.Mutex, clientID string) {
	_, err := channel.SubscribeAll(context.Background(), func(msg *ably.Message) {
		timestamp := time.Now().UnixNano()
		tmp := &model.Payload{}
		if err := json.Unmarshal([]byte(msg.Data.(string)), &tmp); err != nil {
			log.Warnf("Couldn't unmarshal msg %v, error: %v", msg.Data, err)
		}
		verifyMessage(tmp, clientID, timestamp, results, mu, channel)
	})
	if err != nil {
		log.Errorf("Couldn't subscribe to channel: %v\n", err)
	}
}

func createMsg(name string, msgNum int) model.Payload {
	return model.Payload{
		Message:              fmt.Sprintf("Sending message from device %s", name),
		MessageNum:           msgNum,
		OriginDevice:         name,
		OriginTimestamp:      time.Now().UnixNano(),
		DestinationDevice:    "",
		DestinationTimestamp: 0,
	}
}
