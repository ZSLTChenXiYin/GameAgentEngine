package workertest

import "testing"

func TestAssertHelpers(t *testing.T) {
	if err := AssertTrue(true, "boom"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := AssertEqual(1, 1, "boom"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := AssertTrue(false, "boom"); err == nil {
		t.Fatal("expected error for false assertion")
	}
	if err := AssertEqual(1, 2, "boom"); err == nil {
		t.Fatal("expected error for unequal assertion")
	}
}
