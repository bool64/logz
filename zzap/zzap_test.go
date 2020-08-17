package zzap_test

import (
	"strconv"
	"testing"

	"github.com/bool64/logz"
	"github.com/bool64/logz/zzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewOption(t *testing.T) {
	zc := zap.NewProductionConfig()
	zz, lo := zzap.NewOption(logz.Config{
		MaxCardinality: 5,
		MaxSamples:     10,
	})
	zc.OutputPaths = nil

	l, err := zc.Build(zz)
	require.NoError(t, err)

	l.Sugar().Warnw("message", "index", 1)

	entries := lo[zap.WarnLevel+1].GetEntries()
	assert.Equal(t, uint64(1), entries[0].Count)
}

func BenchmarkNewOption(b *testing.B) {
	zc := zap.NewProductionConfig()
	zz, _ := zzap.NewOption(logz.Config{
		MaxCardinality: 5,
		MaxSamples:     10,
	})
	zc.OutputPaths = nil

	l, err := zc.Build(zz)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		l.Sugar().Warnw("message"+strconv.Itoa(i%100), "index", i)
	}
}
