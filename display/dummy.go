package display

import (
	"github.com/lucasb-eyer/go-colorful"
)

type DummyDisplay struct{}

func NewDummyDisplay() ColorDisplay {
	return DummyDisplay{}
}

func (d DummyDisplay) SetColor(id int, color colorful.Color) error {
	return nil
}

func (d DummyDisplay) SetSingleColor(color colorful.Color) error {
	return nil
}

func (d DummyDisplay) DisplaySize() int {
	return 1
}

func (d DummyDisplay) Render() {}
