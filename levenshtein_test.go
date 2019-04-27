package main

import (
	"testing"

	"github.com/gonutz/check"
)

func TestMin(t *testing.T) {
	check.Eq(t, min(1), 1)
	check.Eq(t, min(1, 2), 1)
	check.Eq(t, min(2, 1), 1)
	check.Eq(t, min(1, 2, 3), 1)
	check.Eq(t, min(2, 3, 1), 1)
	check.Eq(t, min(3, 1, 2), 1)
	check.Eq(t, min(1, 3, 2), 1)
	check.Eq(t, min(2, 1, 3), 1)
	check.Eq(t, min(3, 2, 1), 1)
}

func TestLevenshteinDistance(t *testing.T) {
	dist := func(a, b string, want int) {
		t.Helper()
		check.Eq(t, editDistance(a, b), want, "a->b")
		check.Eq(t, editDistance(b, a), want, "b->a")
	}
	dist("", "", 0)
	dist("a", "", 1)
	dist("a", "a", 0)
	dist("a", "b", 1)
	dist("a", "ab", 1)
	dist("a", "abc", 2)
	dist("a", "abcd", 3)
}
