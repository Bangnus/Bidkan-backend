package entity

import "testing"

func TestNormalizePagination(t *testing.T) {
	cases := []struct {
		inPage, inLimit     int
		wantPage, wantLimit int
	}{
		{0, 0, 1, 10},
		{-3, 5, 1, 5},
		{2, 200, 2, 100},
		{3, 25, 3, 25},
	}
	for _, tc := range cases {
		p, l := NormalizePagination(tc.inPage, tc.inLimit)
		if p != tc.wantPage || l != tc.wantLimit {
			t.Errorf("NormalizePagination(%d,%d) = (%d,%d), want (%d,%d)", tc.inPage, tc.inLimit, p, l, tc.wantPage, tc.wantLimit)
		}
	}
}

func TestNewPageMeta(t *testing.T) {
	m0 := NewPageMeta(1, 10, 0)
	if m0.Total != 0 || m0.TotalPages != 1 {
		t.Fatalf("total=0 got %+v", m0)
	}
	m25 := NewPageMeta(2, 10, 25)
	if m25.Page != 2 || m25.Limit != 10 || m25.Total != 25 || m25.TotalPages != 3 {
		t.Fatalf("total=25 got %+v", m25)
	}
}
