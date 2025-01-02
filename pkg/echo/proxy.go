package echo

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/labstack/gommon/log"
)

// Proxy is a proxy type for logrus.Logger
type Proxy struct {
	dest *slog.Logger
}

var proxy *Proxy

// NewProxyFor creates a new Proxy for a particular logger
func NewProxyFor(logger *slog.Logger) *Proxy {
	return &Proxy{
		dest: logger.With(slog.Bool("echo", true)),
	}
}

// NewProxy creates a new Proxy for the standard logger
func NewProxy() *Proxy {
	return NewProxyFor(slog.Default())
}

func SetDefault(p *Proxy) {
	proxy = p
}

var _ = Logger(&Proxy{})

func (p *Proxy) Write(b []byte) (int, error) {
	p.dest.Info(string(b))
	return len(b), nil
}

func (p *Proxy) Output() io.Writer {
	return p
}

func (p *Proxy) SetOutput(w io.Writer) {
	// noop
}

func (p *Proxy) Prefix() string {
	return ""
}

func (p *Proxy) SetPrefix(prefix string) {
	// noop
}

func (p *Proxy) Level() log.Lvl {
	ctx := context.Background()
	if p.dest.Enabled(ctx, slog.LevelDebug) {
		return log.DEBUG
	}

	if p.dest.Enabled(ctx, slog.LevelInfo) {
		return log.INFO
	}

	if p.dest.Enabled(ctx, slog.LevelWarn) {
		return log.WARN
	}

	if p.dest.Enabled(ctx, slog.LevelError) {
		return log.ERROR
	}

	return log.OFF
}

func (p *Proxy) SetLevel(l log.Lvl) {
	// noop
}

func (p *Proxy) SetHeader(h string) {
	// noop
}

func (p *Proxy) Print(args ...interface{}) {
	p.dest.Info(fmt.Sprint(args...))
}

func (p *Proxy) Printf(format string, args ...interface{}) {
	p.dest.Info(fmt.Sprintf(format, args...))
}

func (p *Proxy) Printj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelInfo, msg, args...)
}

func (p *Proxy) Debug(args ...interface{}) {
	p.dest.Debug(fmt.Sprint(args...))
}

func (p *Proxy) Debugf(format string, args ...interface{}) {
	p.dest.Debug(fmt.Sprintf(format, args...))
}

func (p *Proxy) Debugj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelDebug, msg, args...)
}

func (p *Proxy) Info(args ...interface{}) {
	p.dest.Info(fmt.Sprint(args...))
}

func (p *Proxy) Infof(format string, args ...interface{}) {
	p.dest.Info(fmt.Sprintf(format, args...))
}

func (p *Proxy) Infoj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelInfo, msg, args...)
}

func (p *Proxy) Warn(args ...interface{}) {
	p.dest.Warn(fmt.Sprint(args...))
}

func (p *Proxy) Warnf(format string, args ...interface{}) {
	p.dest.Warn(fmt.Sprintf(format, args...))
}

func (p *Proxy) Warnj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelWarn, msg, args...)
}

func (p *Proxy) Error(args ...interface{}) {
	p.dest.Error(fmt.Sprint(args...))
}

func (p *Proxy) Errorf(format string, args ...interface{}) {
	p.dest.Error(fmt.Sprintf(format, args...))
}

func (p *Proxy) Errorj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelError, msg, args...)
}

func (p *Proxy) Fatal(args ...interface{}) {
	p.dest.Error(fmt.Sprint(args...))
}

func (p *Proxy) Fatalf(format string, args ...interface{}) {
	p.dest.Error(fmt.Sprintf(format, args...))
}

func (p *Proxy) Fatalj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelError, msg, args...)
}

func (p *Proxy) Panic(args ...interface{}) {
	p.dest.Error(fmt.Sprint(args...))
}

func (p *Proxy) Panicf(format string, args ...interface{}) {
	p.dest.Error(fmt.Sprintf(format, args...))
}

func (p *Proxy) Panicj(j log.JSON) {
	args := make([]slog.Attr, 0, len(j))
	msg := ""
	for k, v := range j {
		if k == "msg" {
			msg = fmt.Sprint(v)
			continue
		}
		args = append(args, slog.Any(k, v))
	}
	p.dest.LogAttrs(context.Background(), slog.LevelError, msg, args...)
}
