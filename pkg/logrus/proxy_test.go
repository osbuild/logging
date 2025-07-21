package logrus

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"testing"

	"github.com/osbuild/logging/pkg/collect"
)

func TestLogrusProxy(t *testing.T) {
	tests := []struct {
		f      func(*Proxy)
		result map[string]any
	}{
		{
			f: func(p *Proxy) {
				p.Trace("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Debug("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Info("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Warn("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Error("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Fatal("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Panic("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},

		{
			f: func(p *Proxy) {
				p.Tracef("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Debugf("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Infof("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Warnf("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Errorf("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Fatalf("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Panicf("message: %d", 42)
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message: 42",
			},
		},

		{
			f: func(p *Proxy) {
				p.Traceln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Debugln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Infoln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Warnln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Errorln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Fatalln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Panicln("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},

		{
			f: func(p *Proxy) {
				p.WithContext(context.Background()).Trace("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
			},
		},

		{
			f: func(p *Proxy) {
				p.WithField("k1", "v").WithField("k2", 42).Trace("message")
			},
			result: map[string]any{
				"logrus": true,
				"msg":    "message",
				"k1":     "v",
				"k2":     int64(42),
			},
		},
	}

	ch := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(ch)
	p := NewProxyFor(logger, Options{NoExit: true})

	for i, tt := range tests {
		t.Run(fmt.Sprintf("t-%d", i), func(t *testing.T) {
			tt.f(p)
			result := ch.Last()

			if !reflect.DeepEqual(tt.result, result) {
				t.Errorf("unexpected result: %v expected: %v", result, tt.result)
			}
		})
	}
}

func TestLogrusProxyFuncs(t *testing.T) {
	ch := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(ch)
	SetDefault(NewProxyFor(logger, Options{NoExit: true}))
	defer SetDefault(nil)

	Traceln("trace")
	Debugln("debug")
	Infoln("info")
	Warnln("warn")
	Errorln("error")
	Panicln("panic")

	Trace("a", "b", "c")
	Debug("a", "b", "c")
	Info("a", "b", "c")
	Warn("a", "b", "c")
	Error("a", "b", "c")
	Panic("a", "b", "c")

	Tracef("number: %d", 42)
	Debugf("number: %d", 42)
	Infof("number: %d", 42)
	Warnf("number: %d", 42)
	Warningf("number: %d", 42)
	Errorf("number: %d", 42)
	Panicf("number: %d", 42)

	ctx := context.Background()
	WithContext(ctx).Trace("msg with context")
	WithContext(ctx).Debug("msg with context")
	WithContext(ctx).Info("msg with context")
	WithContext(ctx).Warn("msg with context")
	WithContext(ctx).Error("msg with context")
	WithContext(ctx).Panic("msg with context")

	WithField("key", "value").Trace("msg with field")
	WithField("key", "value").Debug("msg with field")
	WithField("key", "value").Info("msg with field")
	WithField("key", "value").Warn("msg with field")
	WithField("key", "value").Error("msg with field")
	WithField("key", "value").Panic("msg with field")

	log := ch.All()
	if len(log) != 31 {
		t.Errorf("unexpected log length: %d", len(log))
	}
}
