package main

import (
	"time"

	"github.com/github/git-lfs/Godeps/_workspace/src/github.com/cheggaaa/pb"
)

func main() {
	count := 5000
	bar := pb.New(count)

	// show percents (by default already true)
	bar.ShowPercent = true

	// show bar (by default already true)
	bar.ShowPercent = true

	// no need counters
	bar.ShowCounters = true

	bar.ShowTimeLeft = true

	// and start
	bar.Start()
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}
	bar.FinishPrint("The End!")
}
