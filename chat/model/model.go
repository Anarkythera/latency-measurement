package model

type Payload struct {
	Message              string
	OriginDevice         string
	OriginTimestamp      int64
	DestinationDevice    string
	DestinationTimestamp int64
}
