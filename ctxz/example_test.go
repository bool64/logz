package ctxz_test

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/bool64/ctxd"
	"github.com/bool64/logz"
	"github.com/bool64/logz/ctxz"
	"github.com/bool64/logz/logzpage"
)

func ExampleNewObserver() {
	var logger ctxd.Logger

	lz := ctxz.NewObserver(logger, logz.Config{
		MaxCardinality:      100,
		MaxSamples:          50,
		DistRetentionPeriod: 72 * time.Hour,
	})
	logger = lz

	ctx := context.TODO()

	logger.Debug(ctx, "starting example")
	logger.Info(ctx, "sample info", "one", 1, "two", 2)
	logger.Error(ctx, "unexpected end of the world")

	logger.Important(ctx, "starting server at http://localhost:6060/")

	err := http.ListenAndServe("0.0.0.0:6060", logzpage.Handler(lz.LevelObservers()...))
	if err != nil {
		logger.Error(ctx, err.Error())
		os.Exit(1)
	}
}
