package model

import (
	"errors"
	"strings"
	"testing"
)

func TestHashError(t *testing.T) {
	inner := errors.New("boom")
	err := NewHashError("/tmp/a.txt", inner)
	if !strings.Contains(err.Error(), "hash") || !strings.Contains(err.Error(), "/tmp/a.txt") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(err, inner) {
		t.Fatal("Unwrap should expose inner error")
	}
}

func TestMoveError(t *testing.T) {
	inner := errors.New("denied")
	err := NewMoveError("/src", "/dst", inner)
	msg := err.Error()
	if !strings.Contains(msg, "/src") || !strings.Contains(msg, "/dst") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(err, inner) {
		t.Fatal("Unwrap should expose inner error")
	}
}
