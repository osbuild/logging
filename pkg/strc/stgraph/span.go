package main

import (
	"time"
)

// Span is a span of a trace. See strc package documentation for information
// about the fields.
type Span struct {
	TraceID  string        `json:"trace"`
	ID       string        `json:"id"`
	ParentID string        `json:"parent"`
	Name     string        `json:"name"`
	Source   string        `json:"source"`
	Duration time.Duration `json:"dur"`

	// Children are the child spans of this span.
	Children []*Span

	// Parent is the parent span of this span.
	Parent *Span
}

// TraceWindow is a window of spans that are waiting for their parent to be added.
func (s *Span) Each(f func(*Span)) {
	f(s)
	for _, c := range s.Children {
		c.Each(f)
	}
}

// EachWithLevel calls the function with the span and its level in the tree.
func (s *Span) EachWithLevel(f func(*Span, int), level int) {
	f(s, level)
	for _, c := range s.Children {
		c.EachWithLevel(f, level+1)
	}
}

// EachWithParent calls the function with the span and its parent.
func (s *Span) EachWithParent(f func(*Span, *Span), parent *Span) {
	f(s, parent)
	for _, c := range s.Children {
		c.EachWithParent(f, s)
	}
}

// JoinNames returns the full name of the span joined by dots. Example:
// "root.child.grandchild".
func (s *Span) JoinNames(sep string) string {
	if s.Parent == nil {
		return s.Name
	}

	return s.Parent.JoinNames(sep) + sep + s.Name
}
