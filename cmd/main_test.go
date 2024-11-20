package main

import (
	"testing"
	"time"
)

func BenchmarkSleep(b *testing.B) {
	tickerPause := (time.Minute / 128) / 4
	time.Sleep(tickerPause)
}
