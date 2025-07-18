package sinit

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	sb := &strings.Builder{}
	logger := log.New(sb, "test: ", log.LstdFlags)
	writer := &logLoggerWriter{dest: logger}

	message := "A message"
	n, err := writer.Write([]byte(message))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logOutput := sb.String()
	if !strings.Contains(logOutput, message) {
		t.Fatalf("expected log output to contain %q, got %q", message, logOutput)
	}

	if n != len(message) {
		t.Fatalf("expected %d bytes written, got %d", len(message), n)
	}
}

func TestUninitializedWrite(t *testing.T) {
	writer := &logLoggerWriter{}

	_, err := writer.Write([]byte("This should fail"))
	if !errors.Is(err, ErrLoggerNotInitialized) {
		t.Fatalf("expected error %v, got %v", ErrLoggerNotInitialized, err)
	}
}

func TestCallDepth(t *testing.T) {
	sb := &strings.Builder{}
	logger := log.New(sb, "test: ", log.Llongfile|log.LstdFlags)
	writer := &logLoggerWriter{dest: logger}

	callLineNumber := 48 // Keep it up to date with the line number in the test function
	message := "Depths of Hyrule"
	n, err := writer.Write([]byte(message))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if n != len(message) {
		t.Fatalf("expected %d bytes written, got %d", len(message), n)
	}

	logOutput := sb.String()
	if !strings.Contains(logOutput, message) {
		t.Fatalf("expected log output to contain %q, got %q", message, logOutput)
	}

	if !strings.Contains(logOutput, fmt.Sprintf("%s:%d", "writer_test.go", callLineNumber)) {
		t.Fatalf("expected log output to contain call depth information, got %q", logOutput)
	}
}
