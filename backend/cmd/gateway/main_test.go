package main

import "testing"

func TestPaginationBoundsNoParamsReturnsFullRange(t *testing.T) {
	start, end := paginationBounds(5, "", "")
	if start != 0 || end != 5 {
		t.Errorf("expected [0,5), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsOffsetAndLimit(t *testing.T) {
	start, end := paginationBounds(10, "3", "2")
	if start != 3 || end != 5 {
		t.Errorf("expected [3,5), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsLimitClampedToTotal(t *testing.T) {
	start, end := paginationBounds(4, "2", "100")
	if start != 2 || end != 4 {
		t.Errorf("expected [2,4), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsOffsetBeyondTotal(t *testing.T) {
	start, end := paginationBounds(4, "50", "10")
	if start != 4 || end != 4 {
		t.Errorf("expected empty range [4,4), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsInvalidParamsIgnored(t *testing.T) {
	start, end := paginationBounds(5, "not-a-number", "also-not-a-number")
	if start != 0 || end != 5 {
		t.Errorf("expected fallback to full range [0,5), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsNegativeParamsIgnored(t *testing.T) {
	start, end := paginationBounds(5, "-1", "-1")
	if start != 0 || end != 5 {
		t.Errorf("expected fallback to full range [0,5) for negative params, got [%d,%d)", start, end)
	}
}
