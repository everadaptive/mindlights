package controller

import (
	"github.com/dnlo/struct2csv"
	"gobot.io/x/gobot/platforms/neurosky"
)

type EEGFullData struct {
	Timestamp string
	neurosky.FullData
}

func (d EEGFullData) GetHeaders() []string {
	enc := struct2csv.New()
	values, _ := enc.GetColNames(d)

	return values
}

func (d EEGFullData) ToSlice() []string {
	enc := struct2csv.New()
	values, _ := enc.GetRow(d)

	return values
}
