package logz_test

import (
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/bool64/logz"
	"github.com/stretchr/testify/assert"
)

func TestObserver_ObserveMessage(t *testing.T) {
	o := logz.NewObserver(logz.Config{})

	o.ObserveMessage("test", 123)
	o.ObserveMessage("test", 456)
	o.ObserveMessage("another test", 789)

	entries := o.GetEntries()

	// Order of entries may be random as they are retrieved from a map.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Message > entries[j].Message
	})

	assert.Equal(t, "test", entries[0].Message)
	assert.Equal(t, uint64(2), entries[0].Count)
	assert.NotEmpty(t, entries[0].First)
	assert.NotEmpty(t, entries[0].Last)

	assert.Equal(t, "another test", entries[1].Message)
	assert.Equal(t, uint64(1), entries[1].Count)
	assert.NotEmpty(t, entries[1].First)
	assert.NotEmpty(t, entries[1].Last)

	entry := o.Find("test")
	assert.Equal(t, "test", entry.Message)
	assert.Equal(t, uint64(2), entry.Count)
	assert.NotEmpty(t, entry.First)
	assert.NotEmpty(t, entry.Last)
	assert.NotEmpty(t, entry.Samples)
}

func TestObserver_ObserveMessage_filter(t *testing.T) {
	o := logz.NewObserver(logz.Config{FilterMessage: true})

	o.ObserveMessage("test foo123", 123)
	o.ObserveMessage("test bar456", 456)
	o.ObserveMessage("another test", 789)

	entries := o.GetEntries()

	// Order of entries may be random as they are retrieved from a map.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Message > entries[j].Message
	})

	assert.Equal(t, "test X", entries[0].Message)
	assert.Equal(t, uint64(2), entries[0].Count)
	assert.NotEmpty(t, entries[0].First)
	assert.NotEmpty(t, entries[0].Last)

	assert.Equal(t, "another test", entries[1].Message)
	assert.Equal(t, uint64(1), entries[1].Count)
	assert.NotEmpty(t, entries[1].First)
	assert.NotEmpty(t, entries[1].Last)

	entry := o.Find("test X")
	assert.Equal(t, "test X", entry.Message)
	assert.Equal(t, uint64(2), entry.Count)
	assert.NotEmpty(t, entry.First)
	assert.NotEmpty(t, entry.Last)
	assert.NotEmpty(t, entry.Samples)
}

func BenchmarkObserver_ObserveMessage(b *testing.B) {
	o := logz.NewObserver(logz.Config{})
	wg := sync.WaitGroup{}
	concurrency := 50

	b.ReportAllocs()

	for i := 0; i < concurrency; i++ {
		msg := "message" + strconv.Itoa(i)

		wg.Add(1)

		go func() {
			for i := 0; i < b.N/concurrency; i++ {
				o.ObserveMessage(msg, i)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkObserver_ObserveMessage_filter(b *testing.B) {
	o := logz.NewObserver(logz.Config{FilterMessage: true})
	wg := sync.WaitGroup{}
	concurrency := 50

	b.ReportAllocs()

	for i := 0; i < concurrency; i++ {
		msg := "foo bar message" + strconv.Itoa(i)

		wg.Add(1)

		go func() {
			for i := 0; i < b.N/concurrency; i++ {
				o.ObserveMessage(msg, i)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
