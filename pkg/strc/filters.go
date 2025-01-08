package strc

import (
	"net/http"
	"regexp"
	"strings"
)

// This code is coming from https://github.com/samber/slog-http

// Filter is a function that determines whether a request should be logged or not.
type Filter func(w WrapResponseWriter, r *http.Request) bool

// Accept returns a filter that accepts requests that pass the given filter.
func Accept(filter Filter) Filter { return filter }

// Ignore returns a filter that ignores requests that pass the given filter.
func Ignore(filter Filter) Filter { return filter }

// AcceptMethod returns a filter that accepts requests with the given methods.
func AcceptMethod(methods ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		reqMethod := strings.ToLower(r.Method)

		for _, method := range methods {
			if strings.ToLower(method) == reqMethod {
				return true
			}
		}

		return false
	}
}

// IgnoreMethod returns a filter that ignores requests with the given methods.
func IgnoreMethod(methods ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		reqMethod := strings.ToLower(r.Method)

		for _, method := range methods {
			if strings.ToLower(method) == reqMethod {
				return false
			}
		}

		return true
	}
}

// AcceptStatus returns a filter that accepts requests with the given status codes.
func AcceptStatus(statuses ...int) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, status := range statuses {
			if status == w.Status() {
				return true
			}
		}

		return false
	}
}

// IgnoreStatus returns a filter that ignores requests with the given status codes.
func IgnoreStatus(statuses ...int) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, status := range statuses {
			if status == w.Status() {
				return false
			}
		}

		return true
	}
}

// AcceptPath returns a filter that accepts requests with the given paths.
func AcceptPath(urls ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, url := range urls {
			if r.URL.Path == url {
				return true
			}
		}

		return false
	}
}

// IgnorePath returns a filter that ignores requests with the given paths.
func IgnorePath(urls ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, url := range urls {
			if r.URL.Path == url {
				return false
			}
		}

		return true
	}
}

// AcceptPathContains returns a filter that accepts requests with paths that contain the given parts.
func AcceptPathContains(parts ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Path, part) {
				return true
			}
		}

		return false
	}
}

// IgnorePathContains returns a filter that ignores requests with paths that contain the given parts.
func IgnorePathContains(parts ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Path, part) {
				return false
			}
		}

		return true
	}
}

// AcceptPathPrefix returns a filter that accepts requests with paths that have the given prefixes.
func AcceptPathPrefix(prefixs ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

// IgnorePathPrefix returns a filter that ignores requests with paths that have the given prefixes.
func IgnorePathPrefix(prefixs ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return false
			}
		}

		return true
	}
}

// AcceptPathSuffix returns a filter that accepts requests with paths that have the given suffixes.
func AcceptPathSuffix(prefixs ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

// IgnorePathSuffix returns a filter that ignores requests with paths that have the given suffixes.
func IgnorePathSuffix(suffixs ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(r.URL.Path, suffix) {
				return false
			}
		}

		return true
	}
}

// AcceptPathMatch returns a filter that accepts requests with paths that match the given regular expressions.
func AcceptPathMatch(regs ...regexp.Regexp) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Path)) {
				return true
			}
		}

		return false
	}
}

// IgnorePathMatch returns a filter that ignores requests with paths that match the given regular expressions.
func IgnorePathMatch(regs ...regexp.Regexp) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Path)) {
				return false
			}
		}

		return true
	}
}

// AcceptHost returns a filter that accepts requests with the given hosts.
func AcceptHost(hosts ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, host := range hosts {
			if r.URL.Host == host {
				return true
			}
		}

		return false
	}
}

// IgnoreHost returns a filter that ignores requests with the given hosts.
func IgnoreHost(hosts ...string) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, host := range hosts {
			if r.URL.Host == host {
				return false
			}
		}

		return true
	}
}

// AcceptHostContains returns a filter that accepts requests with hosts that contain the given regular expression.
func AcceptHostMatch(regs ...regexp.Regexp) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Host)) {
				return true
			}
		}

		return false
	}
}

// IgnoreHostContains returns a filter that ignores requests with hosts that contain the given regular expression.
func IgnoreHostMatch(regs ...regexp.Regexp) Filter {
	return func(w WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Host)) {
				return false
			}
		}

		return true
	}
}
