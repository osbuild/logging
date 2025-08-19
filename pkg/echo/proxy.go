package echo

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/labstack/gommon/log"
)

// Proxy is a proxy type for logrus.Logger
type Proxy struct {
	dest *slog.Logger
	ctx  context.Context
}

var proxy *Proxy

// NewProxyWithContextFor creates a new Proxy for a particular logger with a particular context.
func NewProxyWithContextFor(logger *slog.Logger, ctx context.Context) *Proxy {
	return &Proxy{
		dest: logger.With(slog.Bool("echo", true)),
		ctx:  ctx,
	}
}

// NewProxyFor creates a new Proxy for a particular logger
func NewProxyFor(logger *slog.Logger) *Proxy {
	return NewProxyWithContextFor(logger, context.Background())
}

// NewProxy creates a new Proxy for the standard logger
func NewProxy() *Proxy {
	return NewProxyFor(slog.Default())
}

func SetDefault(p *Proxy) {
	proxy = p
}

func Default() *Proxy {
	return proxy
}

var _ = Logger(&Proxy{})

func (p *Proxy) Write(b []byte) (int, error) {
	p.dest.InfoContext(p.ctx, string(b))
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
	if p.dest.Enabled(p.ctx, slog.LevelDebug) {
		return log.DEBUG
	}

	if p.dest.Enabled(p.ctx, slog.LevelInfo) {
		return log.INFO
	}

	if p.dest.Enabled(p.ctx, slog.LevelWarn) {
		return log.WARN
	}

	if p.dest.Enabled(p.ctx, slog.LevelError) {
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

func anyToString(a []any) []string {
	s := make([]string, 0, len(a))
	for _, v := range a {
		s = append(s, fmt.Sprint(v))
	}
	return s
}

func (p *Proxy) Print(args ...interface{}) {
	p.dest.InfoContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Printf(format string, args ...interface{}) {
	p.dest.InfoContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelInfo, msg, args...)
}

func (p *Proxy) Debug(args ...interface{}) {
	p.dest.DebugContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Debugf(format string, args ...interface{}) {
	p.dest.DebugContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelDebug, msg, args...)
}

func (p *Proxy) Info(args ...interface{}) {
	p.dest.InfoContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Infof(format string, args ...interface{}) {
	p.dest.InfoContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelInfo, msg, args...)
}

func (p *Proxy) Warn(args ...interface{}) {
	p.dest.WarnContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Warnf(format string, args ...interface{}) {
	p.dest.WarnContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelWarn, msg, args...)
}

func (p *Proxy) Error(args ...interface{}) {
	p.dest.ErrorContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Errorf(format string, args ...interface{}) {
	p.dest.ErrorContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelError, msg, args...)
}

func (p *Proxy) Fatal(args ...interface{}) {
	p.dest.ErrorContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Fatalf(format string, args ...interface{}) {
	p.dest.ErrorContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelError, msg, args...)
}

func (p *Proxy) Panic(args ...interface{}) {
	p.dest.ErrorContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Panicf(format string, args ...interface{}) {
	p.dest.ErrorContext(p.ctx, fmt.Sprintf(format, args...))
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
	p.dest.LogAttrs(p.ctx, slog.LevelError, msg, args...)
}
