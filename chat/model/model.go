package model

import "strconv"

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
		l.OriginDevice,
		l.DestinationDevice,
		strconv.FormatInt(l.MinLatency, 10),
		strconv.FormatInt(l.MaxLatency, 10),
		strconv.FormatInt(l.AverageLatency, 10),
	}
}
