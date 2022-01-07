package model

import "fmt"

type Payload struct {
	Message              string
	MessageNum           int
	OriginDevice         string
	OriginTimestamp      int64
	DestinationDevice    string
	DestinationTimestamp int64
}

type CvsLine struct {
	OriginDevice      string
	DestinationDevice string
	MinLatency        int64
	MaxLatency        int64
	AverageLatency    int64
}

func (l CvsLine) ToSlice() []string {
	return []string{
		fmt.Sprintf("Origin Device: %s", l.OriginDevice),
		fmt.Sprintf("Destination Device: %s", l.DestinationDevice),
		fmt.Sprintf("Minimum Latency: %d milliseconds", l.MinLatency),
		fmt.Sprintf("Maximum Latency: %d milliseconds", l.MaxLatency),
		fmt.Sprintf("Average Latency: %d milliseconds", l.AverageLatency),
	}
}
