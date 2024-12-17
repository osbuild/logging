package strc

import (
	"log/slog"
	"slices"
)

// From https://github.com/samber/slog-common/blob/main/groups.go

// appendAttrsToGroup appends newAttrs to the group specified by groups.
func appendAttrsToGroup(groups []string, actualAttrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	actualAttrs = slices.Clone(actualAttrs)

	if len(groups) == 0 {
		return UniqAttrs(append(actualAttrs, newAttrs...))
	}

	for i := range actualAttrs {
		attr := actualAttrs[i]

		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			actualAttrs[i] = slog.Group(groups[0], convertToAny(appendAttrsToGroup(groups[1:], attr.Value.Group(), newAttrs...))...)
			return actualAttrs
		}
	}

	return UniqAttrs(
		append(
			actualAttrs,
			slog.Group(
				groups[0],
				convertToAny(appendAttrsToGroup(groups[1:], []slog.Attr{}, newAttrs...))...,
			),
		),
	)
}

// appendRecordAttrsToAttrs appends record attributes to the attrs making capacity for time
func appendRecordAttrsToAttrs(attrs []slog.Attr, groups []string, record *slog.Record) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs)+record.NumAttrs()+2)
	output = append(output, attrs...)

	slices.Reverse(groups)
	record.Attrs(func(attr slog.Attr) bool {

		for i := range groups {
			attr = slog.Group(groups[i], attr)
		}
		output = append(output, attr)
		return true
	})

	return output
}

func UniqAttrs(attrs []slog.Attr) []slog.Attr {
	return uniqByLast(attrs, func(item slog.Attr) string {
		return item.Key
	})
}

func uniqByLast[T any, U comparable](collection []T, iteratee func(item T) U) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[U]int, len(collection))
	seenIndex := 0

	for _, item := range collection {
		key := iteratee(item)

		if index, ok := seen[key]; ok {
			result[index] = item
			continue
		}

		seen[key] = seenIndex
		seenIndex++
		result = append(result, item)
	}

	return result
}

func convertToAny(attrs []slog.Attr) []any {
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	return anyAttrs
}
