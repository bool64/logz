// Package logz provides events observer.
package logz

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/vearutop/dynhist-go"
	"github.com/vearutop/lograte/filter"
)

// Config defines observer configuration.
type Config struct {
	// Name can be used to identify observer instance in a group, for example a group of log levels.
	Name string

	// MaxCardinality limits number of distinct message families being tracked.
	// All messages that exceed cardinality are grouped together as "other".
	// Default 100.
	MaxCardinality uint32

	// MaxSamples limits a number of latest samples kept for a message family.
	// Default 10.
	MaxSamples uint32

	// SamplingInterval is the minimum amount of time needed to pass from a last Sample collection in particular message family.
	// Messages that are observed too quickly after last sampling are counted, but not sampled.
	// Default 1ms.
	SamplingInterval time.Duration

	// DistResolution is the maximum number of time interval buckets to track distribution in time.
	// Default 100.
	DistResolution int

	// DistRetentionPeriod is maximum age of bucket. Use -1 for unlimited.
	// Default one week (168 hours).
	DistRetentionPeriod time.Duration

	// FilterMessage can reduce cardinality by filtering dynamic parts of messages.
	// It uses github.com/vearutop/lograte/filter.Dynamic
	// See https://pkg.go.dev/github.com/vearutop/lograte/filter#Dynamic.
	// This option is not needed if you already have messages without dynamic interpolated values.
	// This option worsen performance, so use it only if you need it.
	FilterMessage bool
}

// NewObserver creates PreparedObserver.
func NewObserver(cfg Config) *PreparedObserver {
	o := PreparedObserver{}
	o.initialize(cfg)

	return &o
}

// PreparedObserver keeps track of messages.
type PreparedObserver struct {
	samplingInterval    int64
	count               uint32
	maxCardinality      uint32
	maxSamples          uint32
	distResolution      int
	distRetentionPeriod int64
	entries             sync.Map
	other               *entry
	filterMessage       bool
}

// Observer keeps track of messages.
// Use NewObserver() to create new instance.
type Observer struct {
	Config
	PreparedObserver

	once sync.Once
}

type entry struct {
	msg                 string
	samples             chan Sample
	count               uint64
	first               int64
	latest              int64
	distribution        *dynhist.Collector
	distRetentionPeriod int64
}

// Sample is a single sample of a message.
type Sample struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Time time.Time   `json:"time"`
}

func (en *entry) push(now int64, sample Sample) {
	cnt := atomic.AddUint64(&en.count, 1)

	if en.distribution != nil {
		en.distribution.Add(float64(now))

		if en.distRetentionPeriod > 0 {
			en.distribution.Lock()
			if int64(en.distribution.Buckets[0].Min) < now-en.distRetentionPeriod {
				en.distribution.Buckets = append(en.distribution.Buckets[:0:0], en.distribution.Buckets[1:]...)
			}
			en.distribution.Unlock()
		}
	}

	if cnt > uint64(cap(en.samples)) && now <= atomic.LoadInt64(&en.latest) {
		return
	}

	atomic.StoreInt64(&en.latest, now)

	// Push new Sample.
	<-en.samples
	en.samples <- sample
}

func (l *PreparedObserver) initialize(cfg Config) {
	l.samplingInterval = int64(cfg.SamplingInterval)
	if l.samplingInterval == 0 {
		l.samplingInterval = int64(time.Millisecond) // 1ms sampling interval by default.
	}

	l.maxCardinality = cfg.MaxCardinality
	if l.maxCardinality == 0 {
		l.maxCardinality = 100
	}

	l.maxSamples = cfg.MaxSamples
	if l.maxSamples == 0 {
		l.maxSamples = 10
	}

	l.distResolution = cfg.DistResolution
	if l.distResolution == 0 {
		l.distResolution = 100
	}

	l.distRetentionPeriod = int64(cfg.DistRetentionPeriod)
	if l.distRetentionPeriod == 0 {
		l.distRetentionPeriod = int64(168 * time.Hour)
	}

	l.other = &entry{
		samples: make(chan Sample, l.maxSamples),
	}
	for i := uint32(0); i < l.maxSamples; i++ {
		l.other.samples <- Sample{}
	}

	if l.distResolution > 0 {
		l.other.distribution = &dynhist.Collector{
			BucketsLimit: l.distResolution,
		}
		l.other.distRetentionPeriod = l.distRetentionPeriod
	}

	if cfg.FilterMessage {
		l.filterMessage = true
	}
}

// ObserveMessage updates aggregated information about message.
func (l *Observer) ObserveMessage(msg string, data interface{}) {
	l.once.Do(func() {
		l.initialize(l.Config)
	})

	l.PreparedObserver.ObserveMessage(msg, data)
}

// ObserveMessage updates aggregated information about message.
func (l *PreparedObserver) ObserveMessage(msg string, data interface{}) {
	tn := time.Now()
	now := tn.UnixNano() / l.samplingInterval
	s := Sample{
		Msg:  msg,
		Data: data,
		Time: tn,
	}

	if l.filterMessage {
		msg = string(filter.Dynamic([]byte(msg), 200))
	}

	if e, ok := l.entries.Load(msg); ok {
		e.(*entry).push(now, s)

		return
	}

	if atomic.LoadUint32(&l.count) < l.maxCardinality {
		e := entry{
			msg:     msg,
			first:   now,
			count:   0,
			samples: make(chan Sample, l.maxSamples),
		}

		if l.distResolution > 0 {
			e.distribution = &dynhist.Collector{
				BucketsLimit: l.distResolution,
			}
			e.distRetentionPeriod = l.distRetentionPeriod
		}

		for i := uint32(0); i < l.maxSamples; i++ {
			e.samples <- Sample{}
		}
		l.entries.Store(msg, &e)
		atomic.AddUint32(&l.count, 1)

		e.push(now, s)
	} else {
		l.other.push(now, s)
	}
}

func (l *PreparedObserver) exportEntry(en *entry, withSamples bool) Entry {
	if en == nil {
		return Entry{}
	}

	e := Entry{
		Message: en.msg,
		Count:   atomic.LoadUint64(&en.count),
		First:   unsampleTime(en.first * l.samplingInterval),
		Last:    unsampleTime(atomic.LoadInt64(&en.latest) * l.samplingInterval),
	}

	if en.distribution != nil {
		e.Buckets = make([]Bucket, 0, l.distResolution)

		en.distribution.Lock()

		for _, b := range en.distribution.Buckets {
			e.Buckets = append(e.Buckets, Bucket{
				From:  unsampleTime(int64(b.Min) * l.samplingInterval),
				To:    unsampleTime(int64(b.Max) * l.samplingInterval),
				Count: uint64(b.Count),
			})
		}
		en.distribution.Unlock()
	}

	if withSamples {
		e.Samples = make([]Sample, 0, l.maxSamples)

		for i := int(l.maxSamples) - 1; i >= 0; i-- {
			sample := <-en.samples
			en.samples <- sample

			if !sample.Time.IsZero() {
				e.Samples = append(e.Samples, sample)
			}
		}
	}

	return e
}

// Entry contains aggregated information about message.
type Entry struct {
	Message string
	Count   uint64
	Samples []Sample
	First   time.Time
	Last    time.Time

	MaxBucketCount int
	Buckets        []Bucket
}

// Bucket contains count of events in time interval.
type Bucket struct {
	From  time.Time
	To    time.Time
	Count uint64
}

// GetEntries returns a list of observed event entries without data samples.
func (l *PreparedObserver) GetEntries() []Entry {
	result := make([]Entry, 0, atomic.LoadUint32(&l.count))

	l.entries.Range(func(_, value interface{}) bool {
		result = append(result, l.exportEntry(value.(*entry), false))

		return true
	})

	return result
}

// GetEntriesWithSamples returns a list of observed event entries with data samples.
func (l *PreparedObserver) GetEntriesWithSamples() []Entry {
	result := make([]Entry, 0, atomic.LoadUint32(&l.count))

	l.entries.Range(func(_, value interface{}) bool {
		result = append(result, l.exportEntry(value.(*entry), true))

		return true
	})

	return result
}

// Find lookups entry by message.
func (l *PreparedObserver) Find(msg string) Entry {
	var e Entry

	l.entries.Range(func(_, value interface{}) bool {
		if value.(*entry).msg == msg {
			e = l.exportEntry(value.(*entry), true)

			return false
		}

		return true
	})

	return e
}

// Other returns entry for other events.
func (l *PreparedObserver) Other(withSamples bool) Entry {
	return l.exportEntry(l.other, withSamples)
}

func unsampleTime(ns int64) time.Time {
	return time.Unix(ns/1e9, ns%1e9)
}
