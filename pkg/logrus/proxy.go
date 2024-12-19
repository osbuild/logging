package logrus

import (
	"context"
	"fmt"
	"log/slog"
)

// Proxy is a proxy type for logrus.Logger
type Proxy struct {
	dest *slog.Logger
	ctx  context.Context
}

var proxy *Proxy

// NewProxy creates a new Proxy
func NewProxyFor(logger *slog.Logger) *Proxy {
	return &Proxy{
		dest: logger.With(slog.Bool("logrus", true)),
	}
}

func NewProxy() *Proxy {
	return NewProxyFor(slog.Default())
}

func SetDefault(p *Proxy) {
	proxy = p
}

func (p *Proxy) WithContext(ctx context.Context) *Proxy {
	p.ctx = ctx

	return p
}

// WithField converts all values to strings
func (p *Proxy) WithField(key string, value any) *Proxy {
	return &Proxy{
		dest: p.dest.With(slog.String(key, fmt.Sprint(value))),
		ctx:  p.ctx,
	}
}

func (p *Proxy) Trace(args ...any) {
	if p.ctx != nil {
		p.dest.Debug(fmt.Sprint(args...))
	} else {
		p.dest.DebugContext(p.ctx, fmt.Sprint(args...))
	}
}

func (p *Proxy) Debug(args ...any) {
	if p.ctx != nil {
		p.dest.Debug(fmt.Sprint(args...))
	} else {
		p.dest.DebugContext(p.ctx, fmt.Sprint(args...))
	}
}

func (p *Proxy) Info(args ...any) {
	if p.ctx != nil {
		p.dest.Info(fmt.Sprint(args...))
	} else {
		p.dest.InfoContext(p.ctx, fmt.Sprint(args...))
	}
}

func (p *Proxy) Warn(args ...any) {
	if p.ctx != nil {
		p.dest.Warn(fmt.Sprint(args...))
	} else {
		p.dest.WarnContext(p.ctx, fmt.Sprint(args...))
	}
}

func (p *Proxy) Error(args ...any) {
	if p.ctx != nil {
		p.dest.Error(fmt.Sprint(args...))
	} else {
		p.dest.ErrorContext(p.ctx, fmt.Sprint(args...))
	}
}

func (p *Proxy) Fatal(args ...any) {
	p.Error(args...)
}

func (p *Proxy) Panic(args ...any) {
	p.Error(args...)
}

func (p *Proxy) Tracef(format string, args ...any) {
	p.Debug(fmt.Sprintf(format, args...))
}

func (p *Proxy) Debugf(format string, args ...any) {
	p.Debug(fmt.Sprintf(format, args...))
}

func (p *Proxy) Infof(format string, args ...any) {
	p.Info(fmt.Sprintf(format, args...))
}

func (p *Proxy) Warnf(format string, args ...any) {
	p.Warn(fmt.Sprintf(format, args...))
}

func (p *Proxy) Errorf(format string, args ...any) {
	p.Error(fmt.Sprintf(format, args...))
}

func (p *Proxy) Fatalf(format string, args ...any) {
	p.Fatal(fmt.Sprintf(format, args...))
}

func (p *Proxy) Panicf(format string, args ...any) {
	p.Panic(fmt.Sprintf(format, args...))
}

func Trace(args ...any) {
	proxy.Trace(args...)
}

func Debug(args ...any) {
	proxy.Debug(args...)
}

func Info(args ...any) {
	proxy.Info(args...)
}

func Warn(args ...any) {
	proxy.Warn(args...)
}

func Error(args ...any) {
	proxy.Error(args...)
}

func Fatal(args ...any) {
	proxy.Fatal(args...)
}

func Panic(args ...any) {
	proxy.Panic(args...)
}

func Tracef(format string, args ...any) {
	proxy.Tracef(format, args...)
}

func Debugf(format string, args ...any) {
	proxy.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	proxy.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	proxy.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	proxy.Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	proxy.Fatalf(format, args...)
}

func Panicf(format string, args ...any) {
	proxy.Panicf(format, args...)
}

func WithField(key string, value any) *Proxy {
	return proxy.WithField(key, value)
}

func WithContext(ctx context.Context) *Proxy {
	return proxy.WithContext(ctx)
}
