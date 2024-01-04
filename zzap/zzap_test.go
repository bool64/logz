package zzap_test

import (
	"encoding/json"
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

	l.With(zap.String("k", "v")).Sugar().Warnw("message", "index", 1)

	entries := lo[zap.WarnLevel+1].GetEntriesWithSamples()
	assert.Equal(t, uint64(1), entries[0].Count)
	assert.Equal(t, "message", entries[0].Message)
	j, err := json.Marshal(entries[0].Samples[0])
	require.NoError(t, err)
	assert.Contains(t, string(j), `"msg":"message","index":1,"k":"v"`)
}

func BenchmarkLogzSugarWarn(b *testing.B) {
	b.ReportAllocs()

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

func BenchmarkRawSugarWarn(b *testing.B) {
	b.ReportAllocs()

	zc := zap.NewProductionConfig()
	zc.OutputPaths = nil

	l, err := zc.Build()
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		l.Sugar().Warnw("message"+strconv.Itoa(i%100), "index", i)
	}
}
