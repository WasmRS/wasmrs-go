package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"

	wf "github.com/nanobus/iota/go/transport/wasmrs/example/wordfrequency"
)

func main() {
	ctx := context.Background()
	done := make(chan struct{}, 1)

	// Create word frequency component.
	words := flux.NewProcessor[string]()
	counts := flux.NewProcessor[wf.WordCount]()

	comp := wf.WordFrequency{
		Input: wf.WordFrequencyInputs{
			Words: words,
		},
		Output: wf.WordFrequencyOutputs{
			Counts: counts,
		},
	}
	comp.Process(ctx)

	// Subscribe to words with greater than or equal to
	// 1000 occurrences.
	cancel := false
	counts.Subscribe(flux.Subscribe[wf.WordCount]{
		OnNext: func(next wf.WordCount) {
			if next.Count < 1000 {
				// Cancel stream when count falls below
				// 1000.
				cancel = true
			} else {
				fmt.Printf("%6d %s\n", next.Count, next.Word)
			}
		},
		Finally: func(_ rx.SignalType) {
			close(done)
		},
		OnRequest: func(sub rx.Subscription) {
			if cancel {
				sub.Cancel()
			} else {
				sub.Request(100)
			}
		},
	})

	if err := streamInWords(words); err != nil {
		panic(err)
	}

	<-done
}

func streamInWords(words flux.Sink[string]) error {
	// Stream in words from all of shakespeare's works.
	file, err := os.Open("shakespeare.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		words.Next(strings.Trim(scanner.Text(), ","))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Triggers word counts to start streaming.
	words.Complete()

	return nil
}
