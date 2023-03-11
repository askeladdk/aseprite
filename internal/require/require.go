package require

import "testing"

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal("unexpected error:", err)
	}
}

func True(t *testing.T, test bool, args ...any) {
	if !test {
		t.Helper()
		t.Fatal(args...)
	}
}
