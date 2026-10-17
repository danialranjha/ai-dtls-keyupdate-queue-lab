package main

import (
	"encoding/json"
	"fmt"
	"os"
	"queue-lab/coalesce_only"
	"queue-lab/fixture"
	"queue-lab/merged"
	"queue-lab/original"
	"queue-lab/progress_only"
)

func runAll() []fixture.Result {
	out := []fixture.Result{}
	for _, f := range fixture.Cases {
		for _, run := range []func(fixture.Case) fixture.Result{original.Run, coalesce_only.Run, progress_only.Run, merged.Run} {
			out = append(out, run(f))
		}
	}
	return out
}
func main() {
	rows := runAll()
	data := struct {
		Scope string           `json:"scope"`
		Cases int              `json:"fixture_count"`
		Modes int              `json:"mode_count"`
		Rows  []fixture.Result `json:"rows"`
	}{"Synthetic scheduler events; exact Pion methods; captured command starts, no network or cryptography", len(fixture.Cases), 4, rows}
	b, e := json.MarshalIndent(data, "", "  ")
	if e != nil {
		panic(e)
	}
	os.MkdirAll("artifacts", 0755)
	if e = os.WriteFile("artifacts/run.json", append(b, '\n'), 0644); e != nil {
		panic(e)
	}
	s := "# Exact scheduling observations\n\nEight synthetic event schedules, four source combinations. Starts are capture events, not transmitted writes. Counts are not performance or attack-rate estimates.\n\n| Case | Method | Required responses queued | First-drive write starts | First-drive required responses still queued |\n|---|---|---:|---:|---:|\n"
	for _, r := range rows {
		a, b := r.Snapshots[0], r.Snapshots[1]
		writes := 0
		for _, v := range b.Started {
			if v == "ordinary-write" {
				writes++
			}
		}
		s += fmt.Sprintf("| %s | %s | %d | %d | %d |\n", r.Case, r.Mode, a.RequiredQueued, writes, b.RequiredQueued)
	}
	s += "\n## Repeated-request trace\n\nOne existing key update awaits ACK. Three accepted peer request events arrive; one ordinary write is queued.\n\n"
	for _, r := range rows {
		if r.Case == "three_requests_waiting" {
			s += fmt.Sprintf("- %s: queue before drive %v; starts at first drive %v; queue after first drive %v; after one ACK starts %v.\n", r.Mode, r.Snapshots[0].Queue, r.Snapshots[1].Started, r.Snapshots[1].Queue, r.Snapshots[2].Started)
		}
	}
	s += "\nThe merged scheduler leaves one required response pending and starts the ordinary write before ACK. Coalescing alone still blocks the write. Progress alone starts it but keeps three queued responses. Full ACK-drain fixtures show eventual completion in both original and merged methods. The test does not measure time or establish whole-transport safety.\n"
	if e = os.WriteFile("artifacts/evidence.md", []byte(s), 0644); e != nil {
		panic(e)
	}
	fmt.Printf("%d fixtures, %d rows saved\n", len(fixture.Cases), len(rows))
}
