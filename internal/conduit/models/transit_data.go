package models

import "time"

type TransitData struct {
	Timestamp time.Time

	Type  string
	Value interface{}
}
