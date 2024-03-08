// Package ctxz implements wrapper to report aggregated messages for github.com/bool64/ctxd.Logger.
package ctxz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/bool64/ctxd"
	"github.com/bool64/logz"
)

// Observer keeps track of logged messages.
type Observer struct {
	debug     *logz.Observer
	info      *logz.Observer
	important *logz.Observer
	warn      *logz.Observer
	error     *logz.Observer
	logger    ctxd.Logger
}

type tuples struct {
	ctx context.Context
	kv  []interface{}
}

func (t tuples) MarshalJSON() ([]byte, error) {
	kv := t.kv[0:len(t.kv):len(t.kv)]

	ctxFields := ctxd.Fields(t.ctx)
	if len(ctxFields) > 0 {
		kv = append(kv, ctxFields...)
	}

	m := make(map[string]interface{}, len(kv))

	var (
		label string
		ok    bool
	)

	for i, l := range kv {
		if label == "" { //nolint:nestif
			label, ok = l.(string)
			if !ok {
				m["malformedFields"] = kv[i:]

				break
			}
		} else {
			if err, ok := l.(error); ok {
				l = err.Error()

				var se ctxd.StructuredError

				if errors.As(err, &se) {
					for k, v := range se.Fields() {
						m[k] = v
					}
				}
			}

			m[label] = l
			label = ""
		}
	}

	b := bytes.Buffer{}
	e := json.NewEncoder(&b)

	e.SetEscapeHTML(false)

	err := e.Encode(m)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Debug logs debug message.
func (o Observer) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	o.debug.ObserveMessage(msg, tuples{ctx: ctx, kv: keysAndValues})
	o.logger.Debug(ctx, msg, keysAndValues...)
}

// Info logs informational message.
func (o Observer) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	o.info.ObserveMessage(msg, tuples{ctx: ctx, kv: keysAndValues})
	o.logger.Info(ctx, msg, keysAndValues...)
}

// Important logs important information.
func (o Observer) Important(ctx context.Context, msg string, keysAndValues ...interface{}) {
	o.important.ObserveMessage(msg, tuples{ctx: ctx, kv: keysAndValues})
	o.logger.Important(ctx, msg, keysAndValues...)
}

// Warn logs a warning.
func (o Observer) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	o.warn.ObserveMessage(msg, tuples{ctx: ctx, kv: keysAndValues})
	o.logger.Warn(ctx, msg, keysAndValues...)
}

// Error logs an error.
func (o Observer) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	o.error.ObserveMessage(msg, tuples{ctx: ctx, kv: keysAndValues})
	o.logger.Error(ctx, msg, keysAndValues...)
}

// LevelObservers returns .
func (o Observer) LevelObservers() []*logz.Observer {
	return []*logz.Observer{o.debug, o.info, o.important, o.warn, o.error}
}

// WithLogger returns a copy of Observer with logger, level buckets remain the same.
func (o Observer) WithLogger(l ctxd.Logger) Observer {
	o.logger = l

	return o
}

// CtxdLogger is a service provider.
func (o Observer) CtxdLogger() ctxd.Logger {
	return o
}

// NewObserver initializes Observer instance.
func NewObserver(logger ctxd.Logger, conf ...logz.Config) Observer {
	o := Observer{
		logger: logger,
	}

	cfg := logz.Config{}

	if len(conf) == 1 {
		cfg = conf[0]
	}

	cfg.Name = "Debug"
	o.debug = &logz.Observer{Config: cfg}
	cfg.Name = "Info"
	o.info = &logz.Observer{Config: cfg}
	cfg.Name = "Important"
	o.important = &logz.Observer{Config: cfg}
	cfg.Name = "Warning"
	o.warn = &logz.Observer{Config: cfg}
	cfg.Name = "Error"
	o.error = &logz.Observer{Config: cfg}

	return o
}
