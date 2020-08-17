package zzap_test

import (
	"net/http"

	"github.com/bool64/logz"
	"github.com/bool64/logz/logzpage"
	"github.com/bool64/logz/zzap"
	"go.uber.org/zap"
)

func ExampleNewOption() {
	zc := zap.NewDevelopmentConfig()
	zz, lo := zzap.NewOption(logz.Config{
		MaxCardinality: 5,
		MaxSamples:     10,
	})

	l, err := zc.Build(zz)
	if err != nil {
		panic(err)
	}

	l.Debug("starting example")
	l.Sugar().Infow("sample info", "one", 1, "two", 2)
	l.Error("unexpected end of the world")

	l.Info("starting server at http://localhost:6060/")

	err = http.ListenAndServe("0.0.0.0:6060", logzpage.Handler(lo...))
	if err != nil {
		l.Fatal(err.Error())
	}
}
