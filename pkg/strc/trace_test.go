package strc

import (
	"bytes"
	"context"
	"log/slog"
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

func TestLogTextHandler(t *testing.T) {
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

	s, ctx := Start(context.Background(), "test")
	check(`time=? level=DEBUG msg="span test started" span.name=test span.id=IvQORsV span.parent=0000000 span.trace=bqzcRlJahlbbBZH span.source=?`)

	s.Event("one")
	check(`time=? level=DEBUG msg="span test event one" span.name=test span.id=IvQORsV span.parent=0000000 span.trace=bqzcRlJahlbbBZH span.event=one span.at=? span.source=?`)

	s.End()
	check(`time=? level=DEBUG msg="span ? finished in ?" span.name=test span.id=IvQORsV span.parent=0000000 span.trace=bqzcRlJahlbbBZH span.dur=? span.source=?`)

	s, _ = Start(ctx, "level1")
	check(`time=? level=DEBUG msg="span level1 started" span.name=level1 span.id=kYcTpgn span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.source=?`)

	s.Event("one")
	check(`time=? level=DEBUG msg="span level1 event one" span.name=level1 span.id=kYcTpgn span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.event=one span.at=? span.source=?`)

	s.End()
	check(`time=? level=DEBUG msg="span ? finished in ?" span.name=level1 span.id=kYcTpgn span.parent=IvQORsV span.trace=bqzcRlJahlbbBZH span.dur=? span.source=?`)
}
