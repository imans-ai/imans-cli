package common

import "testing"

func TestCombined(t *testing.T) {
	payload := Combined([]string{"a", "b"}, 5)
	if payload.Pagination.Count != 5 {
		t.Fatalf("count = %d", payload.Pagination.Count)
	}
	if payload.Pagination.Fetched != 2 {
		t.Fatalf("fetched = %d", payload.Pagination.Fetched)
	}
}
