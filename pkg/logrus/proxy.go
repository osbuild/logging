package logrus

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Fields map[string]any

type Level uint32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

// Proxy is a proxy type for logrus.Logger
type Proxy struct {
	dest *slog.Logger
	ctx  context.Context
}

var proxy *Proxy

// NewProxyFor creates a new Proxy for a particular logger
func NewProxyFor(logger *slog.Logger) *Proxy {
	return &Proxy{
		dest: logger.With(slog.Bool("logrus", true)),
	}
}

// NewProxy creates a new Proxy for the standard logger
func NewProxy() *Proxy {
	return NewProxyFor(slog.Default())
}

func Default() *Proxy {
	return proxy
}

func SetDefault(p *Proxy) {
	proxy = p
}

func (p *Proxy) GetLevel() Level {
	return TraceLevel
}

func (p *Proxy) SetLevel(level Level) {}

func (p *Proxy) WithContext(ctx context.Context) *Proxy {
	p.ctx = ctx

	return p
}

func (p *Proxy) WithField(key string, value any) *Proxy {
	return &Proxy{
		dest: p.dest.With(slog.Any(key, value)),
		ctx:  p.ctx,
	}
}

func (p *Proxy) WithFields(fields Fields) *Proxy {
	a := make([]any, 0, len(fields))
	for k, v := range fields {
		a = append(a, slog.Any(k, v))
	}
	return &Proxy{
		dest: p.dest.With(a...),
		ctx:  p.ctx,
	}
}

func anyToString(a []any) []string {
	s := make([]string, 0, len(a))
	for _, v := range a {
		s = append(s, fmt.Sprint(v))
	}
	return s
}

func (p *Proxy) Trace(args ...any) {
	p.Debug(args...)
}

func (p *Proxy) Debug(args ...any) {
	if p.ctx != nil {
		p.dest.Debug(strings.Join(anyToString(args), " "))
	} else {
		p.dest.DebugContext(p.ctx, strings.Join(anyToString(args), " "))
	}
}

func (p *Proxy) Info(args ...any) {
	if p.ctx != nil {
		p.dest.Info(strings.Join(anyToString(args), " "))
	} else {
		p.dest.InfoContext(p.ctx, strings.Join(anyToString(args), " "))
	}
}

func (p *Proxy) Warn(args ...any) {
	if p.ctx != nil {
		p.dest.Warn(strings.Join(anyToString(args), " "))
	} else {
		p.dest.WarnContext(p.ctx, strings.Join(anyToString(args), " "))
	}
}

func (p *Proxy) Error(args ...any) {
	if p.ctx != nil {
		p.dest.Error(strings.Join(anyToString(args), " "))
	} else {
		p.dest.ErrorContext(p.ctx, strings.Join(anyToString(args), " "))
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

func (p *Proxy) Warningf(format string, args ...any) {
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

func (p *Proxy) Traceln(args ...any) {
	p.Debug(args...)
}

func (p *Proxy) Debugln(args ...any) {
	p.Debug(args...)
}

func (p *Proxy) Infoln(args ...any) {
	p.Info(args...)
}

func (p *Proxy) Warnln(args ...any) {
	p.Warn(args...)
}

func (p *Proxy) Warningln(args ...any) {
	p.Warn(args...)
}

func (p *Proxy) Errorln(args ...any) {
	p.Error(args...)
}

func (p *Proxy) Fatalln(args ...any) {
	p.Fatal(args...)
}

func (p *Proxy) Panicln(args ...any) {
	p.Panic(args...)
}

func GetLevel() Level {
	return proxy.GetLevel()
}

func SetLevel(level Level) {
	proxy.SetLevel(level)
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

func Warningf(format string, args ...any) {
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

func WithFields(fields Fields) *Proxy {
	return proxy.WithFields(fields)
}

func WithContext(ctx context.Context) *Proxy {
	return proxy.WithContext(ctx)
}

func Traceln(args ...any) {
	proxy.Traceln(args...)
}

func Debugln(args ...any) {
	proxy.Debugln(args...)
}

func Infoln(args ...any) {
	proxy.Infoln(args...)
}

func Warnln(args ...any) {
	proxy.Warnln(args...)
}

func Warningln(args ...any) {
	proxy.Warningln(args...)
}

func Errorln(args ...any) {
	proxy.Errorln(args...)
}

func Fatalln(args ...any) {
	proxy.Fatalln(args...)
}

func Panicln(args ...any) {
	proxy.Panicln(args...)
}
