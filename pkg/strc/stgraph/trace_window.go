package main

// TraceWindow is a window of spans that are waiting for their parent to be added.
// WARNING: This structure is not designed for long-running processes. It will
// consume memory indefinitely.
type TraceWindow struct {
	spans map[string][]*Span
}

// Append adds a span to the window. It returns true if the span is a root span.
// Use Pop to retrieve the span and its children and delete it from the window.
func (w *TraceWindow) Append(s *Span) bool {
	if _, ok := w.spans[s.TraceID]; !ok {
		w.spans[s.TraceID] = make([]*Span, 0)
	}

	w.spans[s.TraceID] = append(w.spans[s.TraceID], s)

	if s.ParentID == "" || s.ParentID == "0000000" {
		return true
	}

	return false
}

// Pop retrieves the spans for a trace ID and deletes them from the window.
func (w *TraceWindow) Pop(traceID string) []*Span {
	spans, ok := w.spans[traceID]
	if !ok {
		return nil
	}

	delete(w.spans, traceID)
	return spans
}
