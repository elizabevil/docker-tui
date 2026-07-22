package state

import (
	"testing"
	"time"
)

func TestEventStateGenerationAndCancellation(t *testing.T) {
	var events EventState
	firstContext, firstGeneration := events.Begin()
	_, secondGeneration := events.Begin()
	if firstGeneration == secondGeneration || events.Current(firstGeneration) || !events.Current(secondGeneration) {
		t.Fatalf("generations = %d, %d", firstGeneration, secondGeneration)
	}
	select {
	case <-firstContext.Done():
	default:
		t.Fatal("starting a generation did not cancel the previous context")
	}
}

func TestEventStateCoalescesDirtyResources(t *testing.T) {
	var events EventState
	events.Begin()
	token, scheduled := events.MarkDirty("container")
	if !scheduled {
		t.Fatal("first event did not schedule a flush")
	}
	secondToken, scheduled := events.MarkDirty("image")
	if scheduled || secondToken != token {
		t.Fatalf("second event scheduled duplicate flush: token=%d scheduled=%v", secondToken, scheduled)
	}
	dirty := events.TakeDirty(token)
	if !dirty["container"] || !dirty["image"] || events.FlushScheduled {
		t.Fatalf("dirty resources = %#v", dirty)
	}
}

func TestEventStateReconnectBackoffIsCapped(t *testing.T) {
	var events EventState
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 30 * time.Second, 30 * time.Second}
	for index, expected := range want {
		if delay := events.RecordFailure(); delay != expected {
			t.Fatalf("failure %d delay = %s, want %s", index+1, delay, expected)
		}
	}
}

func TestEventStateRenewContextCancelsSubscriptionWithoutChangingGeneration(t *testing.T) {
	var events EventState
	firstContext, generation := events.Begin()
	secondContext := events.RenewContext()

	select {
	case <-firstContext.Done():
	default:
		t.Fatal("renewing the context did not cancel the active subscription")
	}
	if secondContext == firstContext || !events.Current(generation) {
		t.Fatalf("renewed event state = %#v", events)
	}
}
