package state

import (
	"errors"
	"testing"
)

func TestProcessStateLifecycleAndStaleResult(t *testing.T) {
	var s ProcessState
	s.Open("one", "api")
	if !s.Loading || s.ContainerID != "one" {
		t.Fatalf("open = %#v", s)
	}
	if s.Apply("stale", []string{"PID"}, [][]string{{"1"}}, nil) {
		t.Fatal("stale process result was applied")
	}
	if !s.Apply("one", []string{"PID"}, [][]string{{"1"}, {"2"}}, nil) {
		t.Fatal("current process result was ignored")
	}
	s.Move(5)
	if s.Cursor != 1 {
		t.Fatalf("bounded cursor = %d", s.Cursor)
	}
	if !s.Apply("one", nil, nil, errors.New("container stopped")) || s.Error == "" {
		t.Fatalf("error result = %#v", s)
	}
	s.Close()
	if s.ContainerID != "" || len(s.Rows) != 0 {
		t.Fatalf("close = %#v", s)
	}
}
