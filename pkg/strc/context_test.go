package strc

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"testing"
)

func TestTraceID(t *testing.T) {
	src = rand.NewSource(0)
	var tid TraceID
	var r *http.Request = &http.Request{
		Header: http.Header{},
	}

	if string(EmptyTraceID) != strings.Repeat("0", traceLength) {
		t.Error("EmptyTraceID is not correct")
	}

	if TraceID(tid.String()) != EmptyTraceID {
		t.Error("Uninitialized TraceID is not EmptyTraceID")
	}

	tid = NewTraceID()
	if tid.String() != "bqzcRlJahlbbBZH" {
		t.Errorf("NewTraceID() = %s want something else", tid)
	}

	if len(tid) != traceLength {
		t.Errorf("NewTraceID() = %s want length %d", tid, traceLength)
	}

	if len(EmptyTraceID) != traceLength {
		t.Errorf("EmptyTraceID = %s want length %d", EmptyTraceID, traceLength)
	}

	tid = NewTraceID()
	if tid.String() != "IvQORsVBkYcTpgn" {
		t.Errorf("NewTraceID() = %s want something else", tid)
	}

	ctx := WithTraceID(context.Background(), NewTraceID())
	tid = TraceIDFromContext(ctx)
	if tid.String() != "VIPEcESnuyuufaH" {
		t.Errorf("NewTraceID() = %s want something else", tid)
	}

	ctx = WithTraceID(context.Background(), NewTraceID())
	AddTraceIDHeader(ctx, r)
	tid = TraceIDFromRequest(r)
	if tid.String() != "LOlIxiHprrrvHqD" {
		t.Errorf("NewTraceID() = %s want something else", tid)
	}
}

func TestSpanID(t *testing.T) {
	src = rand.NewSource(0)
	var sid SpanID
	var r *http.Request = &http.Request{
		Header: http.Header{},
	}

	if string(EmptySpanID) != strings.Repeat("0", spanLength)+"."+strings.Repeat("0", spanLength) {
		t.Error("EmptySpanID is not correct")
	}

	if SpanID(sid.String()) != EmptySpanID {
		t.Error("Uninitialized SpanID is not EmptySpanID")
	}

	sid = NewSpanID(context.Background())
	if sid.String() != "0000000.bqzcRlJ" {
		t.Errorf("NewSpanID() = %s want something else", sid)
	}

	if len(sid) != spanLength*2+1 {
		t.Errorf("NewSpanID() = %s want length %d", sid, spanLength*2+1)
	}

	if len(EmptySpanID) != spanLength*2+1 {
		t.Errorf("EmptySpanID = %s want length %d", EmptySpanID, spanLength*2+1)
	}

	sid = NewSpanID(WithSpanID(context.Background(), sid))
	if sid.String() != "bqzcRlJ.ahlbbBZ" {
		t.Errorf("NewSpanID() = %s want something else", sid)
	}

	ctx := WithSpanID(context.Background(), sid)
	sid = NewSpanID(ctx)
	if sid.String() != "ahlbbBZ.IvQORsV" {
		t.Errorf("NewSpanID() = %s want something else", sid)
	}

	sid = NewSpanID(ctx)
	if sid.String() != "ahlbbBZ.kYcTpgn" {
		t.Errorf("NewSpanID() = %s want something else", sid)
	}

	ctx = WithSpanID(ctx, NewSpanID(ctx))
	sid = SpanIDFromContext(ctx)
	if sid.String() != "ahlbbBZ.VIPEcES" {
		t.Errorf("NewSpanID() = %s want something else", sid)
	}

	ctx = WithSpanID(ctx, NewSpanID(ctx))
	AddSpanIDHeader(ctx, r)
	sid = SpanIDFromRequest(r)
	if sid.String() != "VIPEcES.yuufaHI" {
		t.Errorf("NewSpanID() = %s want something else", sid)
	}
}
