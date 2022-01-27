package display

import "github.com/lucasb-eyer/go-colorful"

type ColorDisplay interface {
	SetColor(id int, color colorful.Color) error
	SetSingleColor(color colorful.Color) error
	DisplaySize() int
}
