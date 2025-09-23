package strc

import (
	"bytes"
	"context"
	"log/slog"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestLogTextHandler(t *testing.T) {
	t.Setenv("TZ", "UTC")
	var buf bytes.Buffer

	src = rand.NewSource(0)
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	SetLogger(l)

	var timeReplace = regexp.MustCompile(`time=[^ ]+`)
	var sourceReplace = regexp.MustCompile(`span.source=[^ ]+`)
	var durReplace = regexp.MustCompile(`span.dur=[^ ]+`)
	var atReplace = regexp.MustCompile(`span.at=[^ ]+`)
	var finishedReplace = regexp.MustCompile(`"span \w+ finished in [^"]+"`)

	check := func(want string) {
		t.Helper()
		str := buf.String()

		str = timeReplace.ReplaceAllString(str, `time=?`)
		str = sourceReplace.ReplaceAllString(str, `span.source=?`)
		str = durReplace.ReplaceAllString(str, `span.dur=?`)
		str = atReplace.ReplaceAllString(str, `span.at=?`)
		str = finishedReplace.ReplaceAllString(str, `"span ? finished in ?"`)

		if strings.TrimSpace(str) != want {
			t.Errorf("got:\n%s\nwant:\n%s\n", str, want)
		}
		buf.Reset()
	}

	s, ctx := Start(context.Background(), "test", "arg1", 1)
	check(`time=? level=DEBUG msg="span test started" span.arg1=1 span.name=test span.id=IvQORsV span.parent=0000000 span.trace=bqzcRlJahlbbBZH span.source=?`)

	s.Event("one", "arg2", 2)
	check(`time=? level=DEBUG msg="span test event one" span.arg1=1 span.arg2=2 span.name=test span.id=IvQORsV span.parent=0000000 span.trace=bqzcRlJahlbbBZH span.event=one span.at=? span.source=?`)

	s.End("arg3", 3)
	check(`time=? level=DEBUG msg="span ? finished in ?" span.arg1=1 span.arg3=3 span.name=test span.id=IvQORsV span.parent=0000000 span.trace=bqzcRlJahlbbBZH span.dur=? span.source=?`)

	s, _ = Start(ctx, "level1")
	check(`time=? level=DEBUG msg="span level1 started" span.name=level1 span.id=kYcTpgn span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.source=?`)

	s.Event("one")
	check(`time=? level=DEBUG msg="span level1 event one" span.name=level1 span.id=kYcTpgn span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.event=one span.at=? span.source=?`)

	s.End()
	check(`time=? level=DEBUG msg="span ? finished in ?" span.name=level1 span.id=kYcTpgn span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.dur=? span.source=?`)

	tm, _ := time.Parse(time.RFC3339, "2013-05-13T19:30:00Z")
	s, _ = Start(ctx, "custom", "started", tm)
	check(`time=? level=DEBUG msg="span custom started" span.started=2013-05-13T19:30:00.000Z span.name=custom span.id=VIPEcES span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.source=?`)

	s.Event("one", "at", tm.Add(1*time.Minute))
	if !strings.Contains(buf.String(), "span.at=1m0s") {
		t.Errorf("buffer does not contain at: %s", buf.String())
	}
	check(`time=? level=DEBUG msg="span custom event one" span.started=2013-05-13T19:30:00.000Z span.at=? span.name=custom span.id=VIPEcES span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.event=one span.at=? span.source=?`)

	s.End("finished", tm.Add(10*time.Minute))
	if !strings.Contains(buf.String(), "span.dur=10m0s") {
		t.Errorf("buffer does not contain duration: %s", buf.String())
	}
	check(`time=? level=DEBUG msg="span ? finished in ?" span.started=2013-05-13T19:30:00.000Z span.finished=2013-05-13T19:40:00.000Z span.name=custom span.id=VIPEcES span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.dur=? span.source=?`)
}

func TestFindArgsPairs(t *testing.T) {
	one := 1
	two := 2
	tests := []struct {
		args []any
		key  string
		want *int
	}{
		{[]any{}, "a", nil},
		{[]any{"a"}, "a", nil},
		{[]any{"a", 1}, "a", &one},
		{[]any{"a", "1"}, "a", nil},
		{[]any{"a", 1, "b", 2}, "b", &two},
		{[]any{"a", 1, "b", 2}, "c", nil},
		{[]any{"a", 1, "b"}, "b", nil},
		{[]any{"a", 1, "b", 2, "a", 3}, "a", &one},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := findArgs[int](tt.args, tt.key)
			if (got == nil) != (tt.want == nil) || (got != nil && *got != *tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindArgsAttrs(t *testing.T) {
	tm, _ := time.Parse("2006-01-02T15:04:05Z, MST", "2013-05-13T19:30:00Z, CET")
	tests := []struct {
		args []any
		key  string
		want *time.Time
	}{
		{[]any{}, "a", nil},
		{[]any{slog.String("x", "v")}, "a", nil},
		{[]any{slog.String("a", "v")}, "a", nil},
		{[]any{slog.Time("a", tm)}, "a", &tm},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := findArgs[time.Time](tt.args, tt.key)
			if (got == nil) != (tt.want == nil) || (got != nil && *got != *tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
