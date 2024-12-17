package strc

import "testing"

func TestNewTraceID(t *testing.T) {
	tid := NewTraceID()

	if len(tid) != 16 {
		t.Errorf("len(TraceID()) = %d; want 16", len(tid))
	}
}
