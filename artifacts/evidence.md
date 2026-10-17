# Exact scheduling observations

Eight synthetic event schedules, four source combinations. Starts are capture events, not transmitted writes. Counts are not performance or attack-rate estimates.

| Case | Method | Required responses queued | First-drive write starts | First-drive required responses still queued |
|---|---|---:|---:|---:|
| ordinary_only | original | 0 | 1 | 0 |
| ordinary_only | coalesce_only | 0 | 1 | 0 |
| ordinary_only | progress_only | 0 | 1 | 0 |
| ordinary_only | merged | 0 | 1 | 0 |
| one_request_waiting | original | 1 | 0 | 1 |
| one_request_waiting | coalesce_only | 1 | 0 | 1 |
| one_request_waiting | progress_only | 1 | 1 | 1 |
| one_request_waiting | merged | 1 | 1 | 1 |
| three_requests_waiting | original | 3 | 0 | 3 |
| three_requests_waiting | coalesce_only | 1 | 0 | 1 |
| three_requests_waiting | progress_only | 3 | 1 | 3 |
| three_requests_waiting | merged | 1 | 1 | 1 |
| no_response_requested | original | 0 | 1 | 0 |
| no_response_requested | coalesce_only | 0 | 1 | 0 |
| no_response_requested | progress_only | 0 | 1 | 0 |
| no_response_requested | merged | 0 | 1 | 0 |
| crossed_local | original | 3 | 0 | 3 |
| crossed_local | coalesce_only | 1 | 0 | 1 |
| crossed_local | progress_only | 3 | 1 | 3 |
| crossed_local | merged | 1 | 1 | 1 |
| other_category | original | 1 | 1 | 0 |
| other_category | coalesce_only | 1 | 1 | 0 |
| other_category | progress_only | 1 | 1 | 0 |
| other_category | merged | 1 | 1 | 0 |
| canceled_write | original | 1 | 0 | 1 |
| canceled_write | coalesce_only | 1 | 0 | 1 |
| canceled_write | progress_only | 1 | 0 | 1 |
| canceled_write | merged | 1 | 0 | 1 |
| full_ack_drain | original | 3 | 0 | 3 |
| full_ack_drain | coalesce_only | 1 | 0 | 1 |
| full_ack_drain | progress_only | 3 | 1 | 3 |
| full_ack_drain | merged | 1 | 1 | 1 |

## Repeated-request trace

One existing key update awaits ACK. Three accepted peer request events arrive; one ordinary write is queued.

- original: queue before drive [required-response required-response required-response ordinary-write]; starts at first drive []; queue after first drive [required-response required-response required-response ordinary-write]; after one ACK starts [required-response].
- coalesce_only: queue before drive [required-response ordinary-write]; starts at first drive []; queue after first drive [required-response ordinary-write]; after one ACK starts [required-response ordinary-write].
- progress_only: queue before drive [required-response required-response required-response ordinary-write]; starts at first drive [ordinary-write]; queue after first drive [required-response required-response required-response]; after one ACK starts [ordinary-write required-response].
- merged: queue before drive [required-response ordinary-write]; starts at first drive [ordinary-write]; queue after first drive [required-response]; after one ACK starts [ordinary-write required-response].

The merged scheduler leaves one required response pending and starts the ordinary write before ACK. Coalescing alone still blocks the write. Progress alone starts it but keeps three queued responses. Full ACK-drain fixtures show eventual completion in both original and merged methods. The test does not measure time or establish whole-transport safety.
