package main

import (
	"queue-lab/coalesce_only"
	"queue-lab/fixture"
	"queue-lab/merged"
	"queue-lab/original"
	"queue-lab/progress_only"
	"reflect"
	"testing"
)

func lookup(name string) fixture.Case {
	for _, f := range fixture.Cases {
		if f.Name == name {
			return f
		}
	}
	panic(name)
}
func has(r fixture.Result, step int, s string) bool {
	for _, v := range r.Snapshots[step].Started {
		if v == s {
			return true
		}
	}
	return false
}
func TestWriteProgress(t *testing.T) {
	f := lookup("one_request_waiting")
	if has(original.Run(f), 1, "ordinary-write") || !has(merged.Run(f), 1, "ordinary-write") {
		t.Fatal("original must wait, merged must start")
	}
}
func TestCoalescing(t *testing.T) {
	f := lookup("three_requests_waiting")
	if original.Run(f).Snapshots[0].RequiredQueued != 3 || merged.Run(f).Snapshots[0].RequiredQueued != 1 {
		t.Fatal("expected 3 versus 1")
	}
}
func TestCoalescingAloneDoesNotUnblock(t *testing.T) {
	f := lookup("three_requests_waiting")
	r := coalesce_only.Run(f)
	if r.Snapshots[0].RequiredQueued != 1 || has(r, 1, "ordinary-write") {
		t.Fatal(r)
	}
}
func TestProgressAloneDoesNotCoalesce(t *testing.T) {
	r := progress_only.Run(lookup("three_requests_waiting"))
	if r.Snapshots[0].RequiredQueued != 3 || !has(r, 1, "ordinary-write") {
		t.Fatal(r)
	}
}
func TestOrdinaryTwin(t *testing.T) {
	f := lookup("ordinary_only")
	if !reflect.DeepEqual(original.Run(f).Snapshots, merged.Run(f).Snapshots) {
		t.Fatal("ordinary differs")
	}
	if !has(merged.Run(f), 1, "ordinary-write") {
		t.Fatal("no write")
	}
}
func TestNoResponseRequestedTwin(t *testing.T) {
	f := lookup("no_response_requested")
	if !reflect.DeepEqual(original.Run(f).Snapshots, merged.Run(f).Snapshots) {
		t.Fatal("no-request differs")
	}
}
func TestResponseRetainedUntilAck(t *testing.T) {
	r := merged.Run(lookup("three_requests_waiting"))
	if has(r, 1, "required-response") || !has(r, 2, "required-response") || r.Snapshots[1].RequiredQueued != 1 {
		t.Fatal(r)
	}
}
func TestCrossedLocalUpdate(t *testing.T) {
	r := merged.Run(lookup("crossed_local"))
	if r.Snapshots[0].RequiredQueued != 1 || !has(r, 1, "local-update") || !has(r, 2, "required-response") {
		t.Fatal(r)
	}
}
func TestUnrelatedFlight(t *testing.T) {
	r := merged.Run(lookup("other_category"))
	if !has(r, 1, "ordinary-write") || !has(r, 1, "required-response") {
		t.Fatal(r)
	}
}
func TestCancellation(t *testing.T) {
	f := lookup("canceled_write")
	if has(original.Run(f), 1, "ordinary-write") || has(merged.Run(f), 1, "ordinary-write") {
		t.Fatal("canceled write started")
	}
}
func TestFiniteAckDrain(t *testing.T) {
	f := lookup("full_ack_drain")
	for _, run := range []func(fixture.Case) fixture.Result{original.Run, coalesce_only.Run, progress_only.Run, merged.Run} {
		r := run(f)
		last := len(r.Snapshots) - 1
		if !has(r, last, "ordinary-write") || len(r.Snapshots[last].Queue) != 0 {
			t.Fatal(r)
		}
	}
}
