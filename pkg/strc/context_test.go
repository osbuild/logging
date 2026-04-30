package strc

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestTraceID(t *testing.T) {
	var tid TraceID
	r := &http.Request{
		Header: http.Header{},
	}

	if string(EmptyTraceID) != strings.Repeat("0", traceLength) {
		t.Error("EmptyTraceID is not correct")
	}

	if TraceID(tid.String()) != EmptyTraceID {
		t.Error("Uninitialized TraceID is not EmptyTraceID")
	}

	tid = NewTraceID()
	if len(tid) != traceLength {
		t.Errorf("NewTraceID() = %s want length %d", tid, traceLength)
	}

	if len(EmptyTraceID) != traceLength {
		t.Errorf("EmptyTraceID = %s want length %d", EmptyTraceID, traceLength)
	}

	// Verify all characters are valid letters
	for _, c := range tid {
		if !strings.ContainsRune(letterBytes, c) {
			t.Errorf("NewTraceID() = %s contains invalid character %c", tid, c)
		}
	}

	tid2 := NewTraceID()
	if len(tid2) != traceLength {
		t.Errorf("NewTraceID() = %s want length %d", tid2, traceLength)
	}

	// IDs should be different (extremely unlikely to be the same)
	if tid == tid2 {
		t.Error("NewTraceID() generated duplicate IDs")
	}

	ctx := WithTraceID(context.Background(), tid)
	tidFromCtx := TraceIDFromContext(ctx)
	if tidFromCtx != tid {
		t.Errorf("TraceIDFromContext() = %s want %s", tidFromCtx, tid)
	}

	ctx = WithTraceID(context.Background(), tid)
	AddTraceIDHeader(ctx, r)
	tidFromReq := TraceIDFromRequest(r)
	if tidFromReq != tid {
		t.Errorf("TraceIDFromRequest() = %s want %s", tidFromReq, tid)
	}

	if r.Header.Get(TraceHTTPHeaderName) != tid.String() {
		t.Error("AddTraceIDHeader() did not add the correct header")
	}
}

func TestSpanID(t *testing.T) {
	var sid SpanID
	r := &http.Request{
		Header: http.Header{},
	}

	if string(EmptySpanID) != strings.Repeat("0", spanLength)+"."+strings.Repeat("0", spanLength) {
		t.Error("EmptySpanID is not correct")
	}

	if SpanID(sid.String()) != EmptySpanID {
		t.Error("Uninitialized SpanID is not EmptySpanID")
	}

	sid = NewSpanID(context.Background())
	if len(sid) != spanLength*2+1 {
		t.Errorf("NewSpanID() = %s want length %d", sid, spanLength*2+1)
	}

	// Verify format: parent is EmptySpanID when no parent in context
	if sid.ParentID() != EmptySpanID.ID() {
		t.Errorf("NewSpanID() parent = %s want %s", sid.ParentID(), EmptySpanID.ID())
	}

	if len(EmptySpanID) != spanLength*2+1 {
		t.Errorf("EmptySpanID = %s want length %d", EmptySpanID, spanLength*2+1)
	}

	// Test hierarchical span IDs
	sid1 := NewSpanID(context.Background())
	ctx := WithSpanID(context.Background(), sid1)
	sid2 := NewSpanID(ctx)

	// Child span should have parent's ID as its parent
	if sid2.ParentID() != sid1.ID() {
		t.Errorf("NewSpanID() parent = %s want %s", sid2.ParentID(), sid1.ID())
	}

	// Generate another child from same parent - should have same parent but different ID
	sid3 := NewSpanID(ctx)
	if sid3.ParentID() != sid1.ID() {
		t.Errorf("NewSpanID() parent = %s want %s", sid3.ParentID(), sid1.ID())
	}
	if sid3.ID() == sid2.ID() {
		t.Error("NewSpanID() generated duplicate child IDs")
	}

	// Test grandchild
	ctx2 := WithSpanID(ctx, sid2)
	sid4 := NewSpanID(ctx2)
	if sid4.ParentID() != sid2.ID() {
		t.Errorf("NewSpanID() parent = %s want %s", sid4.ParentID(), sid2.ID())
	}

	ctx = WithSpanID(context.Background(), sid1)
	AddSpanIDHeader(ctx, r)
	sidFromReq := SpanIDFromRequest(r)
	if sidFromReq != sid1 {
		t.Errorf("SpanIDFromRequest() = %s want %s", sidFromReq, sid1)
	}

	if r.Header.Get(SpanHTTPHeaderName) != sid1.String() {
		t.Error("AddSpanIDHeader() did not add the correct header")
	}
}

// TestConcurrentIDGeneration tests that generating IDs concurrently is safe.
// This test will fail with -race if there's a race condition.
func TestConcurrentIDGeneration(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Test concurrent TraceID generation
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				tid := NewTraceID()
				if len(tid) != traceLength {
					t.Errorf("NewTraceID() = %s want length %d", tid, traceLength)
				}
			}
		}()
	}
	wg.Wait()

	// Test concurrent SpanID generation
	wg.Add(numGoroutines)
	ctx := WithSpanID(context.Background(), NewSpanID(context.Background()))
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				sid := NewSpanID(ctx)
				if len(sid) != spanLength*2+1 {
					t.Errorf("NewSpanID() = %s want length %d", sid, spanLength*2+1)
				}
			}
		}()
	}
	wg.Wait()
}
