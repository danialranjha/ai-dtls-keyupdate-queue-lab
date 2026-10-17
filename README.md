# Pion key-update queue comparison

A bounded, offline comparison of scheduling methods before and after [Pion DTLS PR1236](https://github.com/pion/dtls/pull/1236). With one key update awaiting acknowledgement, repeated peer requests can leave ordinary writes waiting in the original scheduler. The merged code coalesces required responses and lets ordinary writes start while the response waits.

This executes **four exact extracted Go function bodies**, with synthetic event schedules and explicit adapters. The sink captures selected command starts. It does **not** send packets, perform encryption, update real keys, run a model, reproduce a whole transport or compare AI scanners.

## Hypothesis

For an active key-update flight, three accepted peer request events and one ordinary application write, the original methods queue three required responses and do not start the write at the first drive. The merged methods queue one required response and start the write. Coalescing alone and scheduling progress alone have different effects.

## Source and setup

The merged PR attributes these unreleased-API issues to an OSS scanner. It explicitly says they do not receive CVEs. This lab does not reproduce the six separately reported CVEs, authentication changes, junk-datagram handling, or the scanner's discovery process.

- Original: [9c02a60bfe8e86f022273039f96ae41634ffcbea](https://github.com/pion/dtls/blob/9c02a60bfe8e86f022273039f96ae41634ffcbea/internal/handshake/post_handshake.go).
- Merged: [79dc6cb5d933375232b55fd8a70684411fc52afe](https://github.com/pion/dtls/blob/79dc6cb5d933375232b55fd8a70684411fc52afe/internal/handshake/post_handshake.go).
- [source.json](source.json) records verified upstream Git-blob hashes, full source hashes, extracted function hashes and every adapter boundary. Upstream code is MIT licensed; see [LICENSE-PION](LICENSE-PION).

Requires locally installed Go 1.22+ and Python 3. No third-party dependencies or downloads. Validated with Go 1.26.4 on macOS/arm64.

```sh
python3 run.py
```

This verifies copied source bodies, runs `go test -json ./...`, then `go run .`, using a disposable build cache and disabling module downloads. It produces [run.json](artifacts/run.json), [checks.json](artifacts/checks.json) and [evidence.md](artifacts/evidence.md). You can also run `python3 verify_source.py`, `go test ./...` and `go run .` directly with your normal Go build cache.

## Threat model and containment

Assume an already-connected peer can supply accepted key-update requests while an earlier update has not been acknowledged. This lab starts **after** record parsing, authentication and protocol-state validation. It does not discover or establish those capabilities. The copied scheduler and fixture driver are trusted. No attacker-controlled strings, credentials, real endpoints, live sockets or unsafe payloads are used.

Only eight finite synthetic event schedules run. An ordinary command is the inert string `ordinary-write`. Required response commands are recorded but never sent. The fixture explicitly removes an active-flight marker to represent an acknowledgement; this is not a real ACK parser. All results are local records.

## Results

Eight selected schedules times four combinations produce 32 rows. Eleven focused tests pass and 16 copied function bodies match their pinned source hashes. These counts are fixture coverage, not independent attacks or measured prevalence.

For the same three-request fixture, before any acknowledgement:

| Method | Required responses queued | Ordinary writes started at first drive |
|---|---:|---:|
| Original | 3 | 0 |
| Coalescing only | 1 | 0 |
| Progress only | 3 | 1 |
| Merged | 1 | 1 |

The merged method still retains one required response. After the explicit acknowledgement event, that response starts. It does not drop all control work or authorize a second simultaneous key-update flight. The ordinary/no-response-requested twins keep working; a canceled write stays unstarted. A queued local update does not suppress the peer's required response. With enough explicit ACK events, the original also eventually starts the write: this finite fixture is **not** proof of permanent deadlock.

## Technical walkthrough

The original `startQueuedPostHandshake` carries `applicationBlocked`. When a required key update is not eligible because another update is in flight, it sets that flag. The subsequent ordinary write then fails `commandEligible`. The merged loop removes that coupling; key-update commands still check their own in-flight category, while application commands remain eligible.

Separately, merged `queueRequiredKeyUpdateResponse` returns early if a required response is already queued. It does not confuse a local, non-required update with a required response. Both versions insert the required response before the first queued application write. The merged loop can skip the temporarily ineligible response and select the write after it.

The four versions combine exact upstream functions:

| Component | Original | Coalescing only | Progress only | Merged |
|---|---|---|---|---|
| Queue/eligibility methods | old | old | new | new |
| Required-response insertion | old | new | old | new |
| Cancellation check | old | new (identical) | old | new (identical) |

The two hybrids are diagnostic ablations, not shipped Pion revisions. Tests require both the ordinary-write distinction and the queued-response distinction, so either half alone cannot masquerade as the full fix. Exact per-step traces, including ACK-drain behavior, are in the run artifact.

### Adapter boundaries

Each package contains `source.go` with unchanged bodies for `startQueuedPostHandshake`, `commandEligible`, `canceledPostHandshakeCommand`, and `queueRequiredKeyUpdateResponse`. Local package/import declarations supply a tiny request-value shim. Reduced types contain only accessed fields. `harness.go` provides a **replacement** `startPostHandshakeCommand` that captures selected commands and marks a key-update flight active. It does not implement Pion's transmission, errors, completion, cryptography or traffic-generation state. Cancellation completion is a no-op adapter. The driver assumes accepted requests and explicitly supplies ACK events. `verify_source.py` checks copied bodies against source.json; external provenance is grounded in its pinned links and independently verified upstream blob hashes.

The upstream patch says application records may use current keys while an update waits. This lab verifies the **scheduling decision only**, not which encryption keys protect real packets. Do not turn captured starts into claims about successful transmission or end-to-end connection safety.

## Limits and supported engineering check

No timings, latency distribution, concurrent workload, packet loss, retransmission, real ACK handling, resource-exhaustion limit, live agent or model were tested. The helper's error path always succeeds, so error propagation is outside the demonstrated result. These eight schedules are not exhaustive protocol-state exploration. An adversary's ability to reach the starting state must be established separately. The applicability to agent infrastructure is a design lesson, not evidence of a voice-agent incident.

For a transport with this scheduling contract, keep only one equivalent required response pending, preserve it until eligible, and test that ordinary writes can still be selected while a previous update awaits acknowledgement. Pair that test with cancellation, crossed local/peer updates and actual transport tests that verify permitted key use and delivery. Use the upstream fixed implementation and validate integration rather than deploying this reduced harness as a transport.
