package neurosky

import (
	"github.com/dnlo/struct2csv"
)

type EEGFullData struct {
	Timestamp     string
	MindflexEvent MindflexEvent
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
