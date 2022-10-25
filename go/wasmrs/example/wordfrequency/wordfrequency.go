package wordfrequency

import (
	"context"
	"sort"
	"strings"

	"github.com/nanobus/iota/go/wasmrs/rx/flux"
)

func (w *WordFrequency) Process(ctx context.Context) {
	counts := make(map[string]uint64)

	w.Input.Words.Subscribe(flux.Subscribe[string]{
		OnNext: func(word string) {
			// Increment word count.
			word = strings.ToLower(word)
			if _, ok := ignoredWords[word]; !ok {
				counts[word] = counts[word] + 1
			}
		},
		OnComplete: func() {
			// Create array from word map.
			wc := makeWordCounts(counts)

			w.Output.Counts.OnSubscribe(flux.OnSubscribe{
				Request: flux.SliceResponse(w.Output.Counts, wc),
			})
		},
	})
}

func makeWordCounts(counts map[string]uint64) []WordCount {
	wc := make([]WordCount, len(counts))
	i := 0
	for word, count := range counts {
		wc[i] = WordCount{word, count}
		i++
	}

	// Sort from highest frequency to lowest.
	sort.Sort(wordsHighestToLowest(wc))
	return wc
}

type wordsHighestToLowest []WordCount

func (a wordsHighestToLowest) Len() int {
	return len(a)
}

func (a wordsHighestToLowest) Less(i, j int) bool {
	return a[i].Count > a[j].Count
}
func (a wordsHighestToLowest) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

var ignoredWords = map[string]struct{}{
	"i":    {},
	"the":  {},
	"and":  {},
	"to":   {},
	"of":   {},
	"a":    {},
	"my":   {},
	"in":   {},
	"you":  {},
	"is":   {},
	"that": {},
}
