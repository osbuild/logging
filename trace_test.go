package strc

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"testing"
)

func TestLogTextHandler(t *testing.T) {
	var buf bytes.Buffer

	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	SetLogger(l)

	check := func(want string) {
		t.Helper()
		re := regexp.MustCompile(`^time=[^ ]+ level=DEBUG ` + want)
		if !re.MatchString(buf.String()) {
			t.Errorf("got: %s want: %s", buf.String(), re)
		}
		t.Log(buf.String())
		buf.Reset()
	}

	s, ctx := StartContext(context.Background(), "test")
	check(`msg="span test started" span\.name=test span\.trace=[^ ]+ span\.id=[^ ]+`)

	s.Event("one")
	check(`msg="span test event one" span\.name=test span\.event=one span.at=[^ ]+ span\.trace=[^ ]+ span\.id=[^ ]+`)

	s.End()
	check(`msg="span test finished" span\.name=test span.duration=[^ ]+ span\.trace=[^ ]+ span\.id=[^ ]+`)

	s, ctx = StartContext(ctx, "level1")
	check(`msg="span level1 started" span\.name=level1 span\.trace=[^ ]+ span\.id=[^ ]+`)

	s.Event("one")
	check(`msg="span level1 event one" span\.name=level1 span\.event=one span.at=[^ ]+ span\.trace=[^ ]+ span\.id=[^ ]+`)

	s.End()
	check(`msg="span level1 finished" span\.name=level1 span.duration=[^ ]+ span\.trace=[^ ]+ span\.id=[^ ]+`)
}
