package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/bool64/logz"
	"github.com/bool64/logz/logzpage"
	"github.com/bool64/logz/zzap"
	lorem "github.com/drhodes/golorem"
	"go.uber.org/zap"
)

func main() {
	zc := zap.NewDevelopmentConfig()
	zz, lo := zzap.NewOption(logz.Config{
		MaxCardinality: 5,
		MaxSamples:     10,
		FilterMessage:  true,
	})

	l, err := zc.Build(zz)
	if err != nil {
		panic(err)
	}

	zc.OutputPaths = nil
	lw, err := zc.Build(zz)
	if err != nil {
		panic(err)
	}

	l.Debug("starting example")
	l.Sugar().Infow("sample info", "one", 1, "two", 2, "special", "<?&>")
	l.Error("unexpected end of the world")

	lorem.Sentence(3, 6)

	for j := 0; j < 50; j++ {
		j := j

		go func() {
			i := 0
			msg := lorem.Sentence(3, 6)

			for {
				i++

				keysAndValues := make([]interface{}, 0, 10)
				keysAndValues = append(keysAndValues, lorem.Word(3, 6), j, lorem.Word(3, 6), i)

				for k := int64(0); k < rand.Int63n(20); k++ {
					keysAndValues = append(keysAndValues, lorem.Word(3, 6), lorem.Word(3, 6))
				}

				lw.Sugar().Warnw(msg+" Dynamic"+strconv.Itoa(i)+".", keysAndValues...)

				time.Sleep(time.Duration(1e9 * rand.Float64()))
			}
		}()
	}

	l.Info("starting server at http://localhost:6060/")
	err = http.ListenAndServe("localhost:6060", logzpage.Handler(lo...))
	if err != nil {
		l.Fatal(err.Error())
	}
}
