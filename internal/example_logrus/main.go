package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/osbuild/logging/pkg/logrus"
)

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	logrus.SetDefault(logrus.NewProxy())

	logrus.Println("print")
	logrus.Traceln("trace")
	logrus.Debugln("debug")
	logrus.Infoln("info")
	logrus.Warnln("warn")
	logrus.Errorln("error")
	logrus.Panicln("panic")

	logrus.Print("a", "b", "c")
	logrus.Trace("a", "b", "c")
	logrus.Debug("a", "b", "c")
	logrus.Info("a", "b", "c")
	logrus.Warn("a", "b", "c")
	logrus.Warning("a", "b", "c")
	logrus.Error("a", "b", "c")
	logrus.Panic("a", "b", "c")

	logrus.Printf("number: %d", 42)
	logrus.Tracef("number: %d", 42)
	logrus.Debugf("number: %d", 42)
	logrus.Infof("number: %d", 42)
	logrus.Warnf("number: %d", 42)
	logrus.Warningf("number: %d", 42)
	logrus.Errorf("number: %d", 42)
	logrus.Panicf("number: %d", 42)

	logrus.WithContext(ctx).Print("msg with context")
	logrus.WithContext(ctx).Trace("msg with context")
	logrus.WithContext(ctx).Debug("msg with context")
	logrus.WithContext(ctx).Info("msg with context")
	logrus.WithContext(ctx).Warn("msg with context")
	logrus.WithContext(ctx).Error("msg with context")
	logrus.WithContext(ctx).Panic("msg with context")

	logrus.WithField("key", "value").Print("msg with field")
	logrus.WithField("key", "value").Trace("msg with field")
	logrus.WithField("key", "value").Debug("msg with field")
	logrus.WithField("key", "value").Info("msg with field")
	logrus.WithField("key", "value").Warn("msg with field")
	logrus.WithField("key", "value").Error("msg with field")
	logrus.WithField("key", "value").Panic("msg with field")
	logrus.WithError(errors.New("test")).Error("msg with err")

	logrus.WithFields(logrus.Fields{
		"str":   "value",
		"int":   42,
		"bool":  true,
		"float": 3.14,
	}).Trace("msg with fields")
}
