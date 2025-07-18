package logrus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

// Proxy is a proxy type for logrus.Logger
type Proxy struct {
	dest        *slog.Logger
	ctx         context.Context
	exitOnFatal bool
}

var proxy atomic.Pointer[Proxy]

func init() {
	proxy.Store(NewDiscardProxy())
}

// NewProxyFor creates a new Proxy for a particular logger. When exitOnFatal is true, the logger will exit
// the process on fatal errors and panic on panic calls.
func NewProxyFor(logger *slog.Logger, exitOnFatal bool) *Proxy {
	return &Proxy{
		dest:        logger.With(slog.Bool("logrus", true)),
		ctx:         context.Background(),
		exitOnFatal: exitOnFatal,
	}
}

// NewProxy creates a new Proxy for the standard logger. It does not exit on fatal
// errors. To do that, use NewProxyFor with exitOnFatal set to true.
func NewProxy() *Proxy {
	return NewProxyFor(slog.Default(), false)
}

// NewDiscardProxy creates a new Proxy which discards all logs. This is the default logger when not set.
func NewDiscardProxy() *Proxy {
	return NewProxyFor(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})), false)
}

// NewEntry creates a new Proxy for the standard logger. Proxy must be passed as an argument.
func NewEntry(fl FieldLogger) *Proxy {
	return fl.(*Proxy)
}

// StandardLogger returns a standard logger proxy.
func StandardLogger() *Proxy {
	return NewProxy()
}

var (
	_ Logger      = &Proxy{}
	_ StdLogger   = &Proxy{}
	_ FieldLogger = &Proxy{}
)

func Default() *Proxy {
	return proxy.Load()
}

func SetDefault(p *Proxy) {
	proxy.Store(p)
}

func (p *Proxy) GetLevel() Level {
	return TraceLevel
}

func (p *Proxy) SetLevel(level Level) {}

func (p *Proxy) WithContext(ctx context.Context) *Proxy {
	return &Proxy{
		dest: p.dest,
		ctx:  ctx,
	}
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

func (p *Proxy) WithError(err error) *Proxy {
	return &Proxy{
		dest: p.dest.With(slog.String("err", err.Error())),
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

func (p *Proxy) Print(args ...any) {
	p.Info(args...)
}

func (p *Proxy) Trace(args ...any) {
	p.Debug(args...)
}

func (p *Proxy) Debug(args ...any) {
	p.dest.DebugContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Info(args ...any) {
	p.dest.InfoContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Warn(args ...any) {
	p.dest.WarnContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Warning(args ...any) {
	p.dest.WarnContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Error(args ...any) {
	p.dest.ErrorContext(p.ctx, strings.Join(anyToString(args), " "))
}

func (p *Proxy) Fatal(args ...any) {
	p.Error(args...)
	if p.exitOnFatal {
		os.Exit(1)
	}
}

func (p *Proxy) Panic(args ...any) {
	p.Error(args...)
	if p.exitOnFatal {
		panic(strings.Join(anyToString(args), " "))
	}
}

func (p *Proxy) Tracef(format string, args ...any) {
	p.Debug(fmt.Sprintf(format, args...))
}

func (p *Proxy) Printf(format string, args ...any) {
	p.Info(fmt.Sprintf(format, args...))
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

func (p *Proxy) Println(args ...any) {
	p.Info(args...)
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
	return proxy.Load().GetLevel()
}

func SetLevel(level Level) {
	proxy.Load().SetLevel(level)
}

func Print(args ...any) {
	proxy.Load().Print(args...)
}

func Trace(args ...any) {
	proxy.Load().Trace(args...)
}

func Debug(args ...any) {
	proxy.Load().Debug(args...)
}

func Info(args ...any) {
	proxy.Load().Info(args...)
}

func Warn(args ...any) {
	proxy.Load().Warn(args...)
}

func Warning(args ...any) {
	proxy.Load().Warning(args...)
}

func Error(args ...any) {
	proxy.Load().Error(args...)
}

func Fatal(args ...any) {
	proxy.Load().Fatal(args...)
}

func Panic(args ...any) {
	proxy.Load().Panic(args...)
}

func Printf(format string, args ...any) {
	proxy.Load().Printf(format, args...)
}

func Tracef(format string, args ...any) {
	proxy.Load().Tracef(format, args...)
}

func Debugf(format string, args ...any) {
	proxy.Load().Debugf(format, args...)
}

func Infof(format string, args ...any) {
	proxy.Load().Infof(format, args...)
}

func Warnf(format string, args ...any) {
	proxy.Load().Warnf(format, args...)
}

func Warningf(format string, args ...any) {
	proxy.Load().Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	proxy.Load().Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	proxy.Load().Fatalf(format, args...)
}

func Panicf(format string, args ...any) {
	proxy.Load().Panicf(format, args...)
}

func WithField(key string, value any) *Proxy {
	return proxy.Load().WithField(key, value)
}

func WithFields(fields Fields) *Proxy {
	return proxy.Load().WithFields(fields)
}

func WithError(err error) *Proxy {
	return proxy.Load().WithError(err)
}

func WithContext(ctx context.Context) *Proxy {
	return proxy.Load().WithContext(ctx)
}

func Println(args ...any) {
	proxy.Load().Println(args...)
}

func Traceln(args ...any) {
	proxy.Load().Traceln(args...)
}

func Debugln(args ...any) {
	proxy.Load().Debugln(args...)
}

func Infoln(args ...any) {
	proxy.Load().Infoln(args...)
}

func Warnln(args ...any) {
	proxy.Load().Warnln(args...)
}

func Warningln(args ...any) {
	proxy.Load().Warningln(args...)
}

func Errorln(args ...any) {
	proxy.Load().Errorln(args...)
}

func Fatalln(args ...any) {
	proxy.Load().Fatalln(args...)
}

func Panicln(args ...any) {
	proxy.Load().Panicln(args...)
}
