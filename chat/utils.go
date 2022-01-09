package chat

import (
	"encoding/csv"
	"fmt"
	"latencychecker/chat/model"
	"os"
	"sync"

	"github.com/ably/ably-go/ably"
)

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
	title := []string{"Origin Host", "Destination Host", "Minimum Latency (ms)", "Maximum Latency (ms)", "Average Latency (ms)"}
	w := csv.NewWriter(file)

	defer w.Flush()
	// Create column names in first line
	err := w.Write(title)
	if err != nil {
		log.Errorf("Error while writing to file: %v", err)

		return err
	}

	// Process results in to each column, max,min,average of timestamps
	cvs := processLatency(results)

	for _, line := range cvs {
		err = w.Write(line.ToSlice())
		if err != nil {
			log.Warnf("Error writing line to file, %v", err)
		}
	}

	return err
}

func processLatency(results map[string]map[string][]model.Payload) []model.CvsLine {
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

// Without the verification origin == clientID everyone gets every result except when they are the same destiniation.
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
