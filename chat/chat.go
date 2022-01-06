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
)

func Start(apiKey, ch, name, fileLocation string, messageNum, delay, wait int) {
	fmt.Printf("Starting latencyChecker for client %s\n", name)
	fmt.Println("Wait: ", wait)
	fmt.Println("Delay: ", delay)

	// Create mutex to write results
	mu := &sync.Mutex{}

	results := []model.Payload{}

	//create ably channel
	client, err := ably.NewRealtime(
		ably.WithKey(apiKey),
		ably.WithClientID(name),
		ably.WithEchoMessages(true),
	)

	if err != nil {
		fmt.Println("Couldn't create ably client")
	}

	channel := client.Channels.Get(ch)

	//subscribe to newly created channel
	subscribe(channel, &results, mu)

	for i := 0; i < messageNum-1; i++ {
		publishing(channel, createMsg(fmt.Sprintf("%s_%d", name, i)))
		fmt.Println("sleeping for ", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	publishing(channel, createMsg(fmt.Sprintf("%s_%d", name, messageNum-1)))
	//Sleep for the wait time, since this is the last message every other must have arrived
	time.Sleep(time.Duration(wait) * time.Second)

	fmt.Println("Main thread waking")

	// Need to check if it's possible to fully terminate the channel before, to avoid concurrency issues
	/*fmt.Println("Detatching channel")
	if err := channel.Detach(context.Background()); err != nil {
		fmt.Println("Error detatching channel")
	}*/
	client.Channels.Release(context.Background(), ch)

	//mu.Lock()
	results = filterTimedoutMgs(results)
	//mu.Unlock()

	//Create file
	/*file, err := os.Create(fmt.Sprintf("%s/%s_%d", fileLocation, "latencyChecker", time.Now()))
	if err != nil {
		fmt.Printf("Couldn't create file at path %s, %v\n", fileLocation, err)
	}

	//err = createResultsTable(file, results)
	if err != nil {
		fmt.Printf("Couldn't write table to file, %v\n", err)
	}*/

	fmt.Printf("Exiting latencyChecker for client %s\n", name)
}

func filterTimedoutMgs(results []model.Payload) []model.Payload {
	for _, elem := range results {
		fmt.Printf("Elem %+v\n", elem)
	}

	//panic("unimplemented")
	return nil
}

func createResultsTable(file *os.File, results []model.Payload) error {
	// Create headers names in first line
	//
	// Process results in to each column, max,min,average of timestamps
	panic("unimplemented")
}

func publishing(channel *ably.RealtimeChannel, msg model.Payload) {
	fmt.Println("Publishing msg ", msg)
	// Publish the message typed in to the Ably Channel
	err := channel.Publish(context.Background(), "message", msg)
	// await confirmation that message was received by Ably
	if err != nil {
		err := fmt.Errorf("publishing to channel: %w", err)
		fmt.Println(err)
	}
}

//TODO: missing verification of destination and such
func subscribe(channel *ably.RealtimeChannel, results *[]model.Payload, mu *sync.Mutex) {
	// Subscribe to messages sent on the channel
	_, err := channel.SubscribeAll(context.Background(), func(msg *ably.Message) {
		fmt.Printf("Received message from %v: '%v with Timestamp: %v'\n", msg.ClientID, msg.Data)
		tmp := &model.Payload{}
		json.Unmarshal([]byte(msg.Data.(string)), &tmp)
		saveMessage(results, tmp, mu)
	})
	if err != nil {
		err := fmt.Errorf("subscribing to channel: %w", err)
		fmt.Println(err)
	}
}

func saveMessage(results *[]model.Payload, tmp *model.Payload, mu *sync.Mutex) {
	fmt.Println("SAVing msf")
	mu.Lock()
	*results = append(*results, *tmp)
	mu.Unlock()
	//fmt.Println("end SAVing msf")
}

func createMsg(name string) model.Payload {
	return model.Payload{
		Message:              fmt.Sprintf("Sending message from device %s", name),
		OriginDevice:         name,
		OriginTimestamp:      time.Now().UnixNano(),
		DestinationDevice:    "",
		DestinationTimestamp: 0,
	}
}
