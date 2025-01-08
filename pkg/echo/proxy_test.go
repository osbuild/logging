package echo

import (
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
				p.Print("a", "b", "c")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "a b c",
			},
		},
		{
			f: func(p *Proxy) {
				p.Debug("message")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Info("message")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Warn("message")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Error("message")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Fatal("message")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message",
			},
		},
		{
			f: func(p *Proxy) {
				p.Panic("message")
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message",
			},
		},

		{
			f: func(p *Proxy) {
				p.Printf("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Debugf("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Infof("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Warnf("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Errorf("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Fatalf("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},
		{
			f: func(p *Proxy) {
				p.Panicf("message: %d", 42)
			},
			result: map[string]any{
				"echo": true,
				"msg":  "message: 42",
			},
		},

		{
			f: func(p *Proxy) {
				p.Printj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
		{
			f: func(p *Proxy) {
				p.Debugj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
		{
			f: func(p *Proxy) {
				p.Infoj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
		{
			f: func(p *Proxy) {
				p.Warnj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
		{
			f: func(p *Proxy) {
				p.Errorj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
		{
			f: func(p *Proxy) {
				p.Fatalj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
		{
			f: func(p *Proxy) {
				p.Panicj(map[string]any{
					"msg": "test",
					"g": map[string]any{
						"k": "v",
					}})
			},
			result: map[string]any{
				"echo": true,
				"msg":  "test",
				"g": map[string]any{
					"k": "v",
				},
			},
		},
	}

	ch := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(ch)
	p := NewProxyFor(logger)
	old := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(old)

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
