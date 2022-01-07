package chat

import (
	"context"
	"encoding/csv"
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

func Start(apiKey, ch, name, fileLocation string, messageNum, delay, wait int) {
	log = logrus.New()
	mu := &sync.Mutex{}
	results := map[string]map[string][]model.Payload{}

	log.SetLevel(logrus.InfoLevel)

	log.Formatter.(*logrus.TextFormatter).FullTimestamp = true

	log.Infof("Starting %s for client %s\n", APPNAME, name)

	client, err := ably.NewRealtime(
		ably.WithKey(apiKey),
		ably.WithClientID(name),
		ably.WithEchoMessages(false),
	)

	if err != nil {
		log.Errorf("Couldn't create ably client, %v\n", err)
		return
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

	//mu.Lock()
	results = filterTimedOutMgs(results, wait)
	//mu.Unlock()
	//printDebug(results)

	//TODO: Create file, convert latency checker to const, converst csv to const var
	fileName := fmt.Sprintf("%s/%s_%s_%d.%s", fileLocation, APPNAME, name, time.Now().Unix(), "csv")
	f, err := os.Create(fileName)

	if err != nil {
		log.Errorf("Couldn't create file at path %s, %v\n", fileLocation, err)
	}

	defer f.Close()

	log.Infof("Created file %s at location %s\n", fileName, fileLocation)

	//mu.Lock()
	if err := createResultsTable(f, results); err != nil {
		log.Infof("Couldn't write table to file, %v\n", err)

		if err := os.Remove(fileName); err != nil {
			log.Errorf("Error while deleting file %s %v", fileName, err)
		}

		return
	}
	//mu.Unlock()

	if err := client.Channels.Release(context.Background(), ch); err != nil {
		log.Errorf("Error cleaning ably resources: %v", err)
	}

	log.Infof("Exiting latencyChecker for client %s\n", name)
}

func printDebug(results map[string]map[string][]model.Payload) {
	for origin, dests := range results {
		fmt.Println("Origin: ", origin)
		for dest, elems := range dests {
			fmt.Println("Destination: ", dest)
			for _, elem := range elems {
				fmt.Printf("Elem: %+v\n", elem)
			}
		}
	}
}

func filterTimedOutMgs(results map[string]map[string][]model.Payload, delay int) map[string]map[string][]model.Payload {
	index := 0

	for _, destDevices := range results {
		for destDevice, msgs := range destDevices {
			for _, msg := range msgs {
				timeDiff := (msg.DestinationTimestamp - msg.OriginTimestamp) / 1000000

				if timeDiff < int64(delay*1000) {
					msgs[index] = msg
					index++
				}
			}

			destDevices[destDevice] = msgs[:index]
			index = 0
		}
	}

	return results
}

func createResultsTable(file *os.File, results map[string]map[string][]model.Payload) error {
	title := []string{"Origin Host", "Destination Host", "Minimum Latency", "Maximum Latency", "Average Lantency"}
	w := csv.NewWriter(file)

	defer w.Flush()
	// Create column names in first line
	if err := w.Write(title); err != nil {
		log.Errorf("Error while writing to file: %v\n", err)

		return err
	}

	// Process results in to each column, max,min,average of timestamps
	cvs := processLatency(results, w)

	for _, line := range cvs {
		if err := w.Write(line.ToSlice()); err != nil {
			log.Warnf("Error writing line to file, %v\n", err)
		}
	}

	return nil
}

func processLatency(results map[string]map[string][]model.Payload, w *csv.Writer) []model.CvsLine {

	lines := []model.CvsLine{}

	for originDevice, dests := range results {
		line := model.CvsLine{
			OriginDevice: originDevice,
		}
		for dest, elems := range dests {
			line.DestinationDevice = dest

			for _, elem := range elems {
				diff := (elem.DestinationTimestamp - elem.OriginTimestamp) / 1000000
				if diff < line.MinLatency || line.MinLatency == 0 {
					line.MinLatency = diff
				}

				if diff > line.MaxLatency || line.MaxLatency == 0 {
					line.MaxLatency = diff
				}

				line.AverageLatency += diff
			}

			if len(elems) == 0 {
				line.AverageLatency = 0
			} else {
				line.AverageLatency /= int64(len(elems))
			}

			lines = append(lines, line)
		}
	}

	return lines
}

func publishing(channel *ably.RealtimeChannel, msg model.Payload) {
	err := channel.Publish(context.Background(), "message", msg)
	if err != nil {
		log.Errorf("Publishing to channel, %v", err)
	}
}

//TODO: missing verification of destination and such
func subscribe(channel *ably.RealtimeChannel, results map[string]map[string][]model.Payload, mu *sync.Mutex, clientID string) {
	// Subscribe to messages sent on the channel
	_, err := channel.SubscribeAll(context.Background(), func(msg *ably.Message) {
		//fmt.Printf("Received message from %v with Timestamp: %v\n", msg.ClientID, msg.Data)
		log.Infof("Received message from %v with Timestamp: %v\n", msg.ClientID, msg.Data)
		timestamp := time.Now().UnixNano()
		tmp := &model.Payload{}
		if err := json.Unmarshal([]byte(msg.Data.(string)), &tmp); err != nil {
			log.Warnf("Couldn't unmarshal msg %v, error: %v\n", msg.Data, err)
		}
		verifyMessage(tmp, clientID, timestamp, results, mu, channel)
	})
	if err != nil {
		err := fmt.Errorf("subscribing to channel: %w", err)
		fmt.Println(err)
	}
}

// without the verification origin == clientID everyone gets every result except when they are the same destiniation
func verifyMessage(tmp *model.Payload, clientID string, timestamp int64, results map[string]map[string][]model.Payload, mu *sync.Mutex, channel *ably.RealtimeChannel) {
	if tmp.DestinationDevice != "" && tmp.DestinationTimestamp != 0 {
		saveMessage(results, tmp, mu)
	}

	if tmp.DestinationDevice == "" || tmp.DestinationTimestamp == 0 {
		tmp.DestinationDevice = clientID
		tmp.DestinationTimestamp = timestamp
		tmp.Message += fmt.Sprintf(", received in device %s", clientID)
		publishing(channel, *tmp)
	}
}

func saveMessage(results map[string]map[string][]model.Payload, tmp *model.Payload, mu *sync.Mutex) {
	mu.Lock()
	if _, ok := results[tmp.OriginDevice]; !ok {
		results[tmp.OriginDevice] = map[string][]model.Payload{}
	}

	results[tmp.OriginDevice][tmp.DestinationDevice] = append(results[tmp.OriginDevice][tmp.DestinationDevice], *tmp)
	mu.Unlock()
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
