package ui

import rl "github.com/gen2brain/raylib-go/raylib"

type Options struct {
	ColorHighlight  rl.Color
	RowHeight       int
	VerticalPadding int
	ScreenWidth     int
	ScreenHeight    int
}

var options *Options

type Window interface {
	Draw()
}

func GetOptions() *Options {
	if options == nil {
		options = &Options{
			ColorHighlight:  rl.Color{255, 127, 10, 33},
			RowHeight:       20,
			VerticalPadding: 10,
			ScreenWidth:     480,
			ScreenHeight:    320,
		}
	}
	return options
}

func SetOptions(opts Options) {
	options = &opts
}
