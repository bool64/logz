package logz_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/bool64/logz"
)

func BenchmarkObserver_ObserveMessage(b *testing.B) {
	o := logz.Observer{}
	wg := sync.WaitGroup{}
	concurrency := 50

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
