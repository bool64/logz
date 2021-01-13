// Package zzap provides zpage observer for "go.uber.org/zap" logger.
package zzap

import (
	"github.com/bool64/logz"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type obCore struct {
	observers []*logz.Observer
	encoder   zapcore.Encoder
	fields    []zapcore.Field

	zapcore.Core
}

type entry struct {
	encoder zapcore.Encoder
	msg     zapcore.Entry
	fields  []zapcore.Field
}

func (e entry) MarshalJSON() ([]byte, error) {
	b, err := e.encoder.EncodeEntry(e.msg, e.fields)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (c obCore) With(fields []zapcore.Field) zapcore.Core {
	if len(fields) == 0 {
		return c
	}

	c.Core = c.Core.With(fields)
	c.fields = append(c.fields[0:len(c.fields):len(c.fields)], fields...)

	return c
}

func (c obCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return c.Core.Check(entry, checkedEntry.AddCore(entry, c))
}

func (c obCore) Write(msg zapcore.Entry, fields []zapcore.Field) error {
	c.observers[msg.Level+1].ObserveMessage(msg.Message, entry{
		encoder: c.encoder,
		msg:     msg,
		fields:  append(fields, c.fields...),
	})

	return nil
}

// NewOption creates zap option with per-level observers.
func NewOption(cfg logz.Config) (zap.Option, []*logz.Observer) {
	var observers []*logz.Observer

	for i := zapcore.DebugLevel; i <= zapcore.FatalLevel; i++ {
		cfg.Name = i.CapitalString()

		observers = append(observers, &logz.Observer{
			Config: cfg,
		})
	}

	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return obCore{
			observers: observers,
			encoder:   zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			Core:      core,
		}
	}), observers
}
