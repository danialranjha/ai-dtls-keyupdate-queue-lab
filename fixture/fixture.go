package fixture

type Case struct {
	Name                                                          string
	Requests, Acks                                                int
	InFlight, OtherFlight, LocalQueued, NotRequested, CancelWrite bool
}

var Cases = []Case{
	{Name: "ordinary_only"},
	{Name: "one_request_waiting", Requests: 1, InFlight: true, Acks: 1},
	{Name: "three_requests_waiting", Requests: 3, InFlight: true, Acks: 1},
	{Name: "no_response_requested", Requests: 3, InFlight: true, NotRequested: true},
	{Name: "crossed_local", Requests: 3, LocalQueued: true, Acks: 1},
	{Name: "other_category", Requests: 1, OtherFlight: true},
	{Name: "canceled_write", Requests: 1, InFlight: true, CancelWrite: true},
	{Name: "full_ack_drain", Requests: 3, InFlight: true, Acks: 4},
}

type Snapshot struct {
	Phase             string   `json:"phase"`
	Queue             []string `json:"queue"`
	Started           []string `json:"captured_starts"`
	RequiredQueued    int      `json:"required_responses_queued"`
	KeyUpdateInFlight bool     `json:"key_update_in_flight"`
}
type Result struct {
	Case      string     `json:"case"`
	Mode      string     `json:"mode"`
	Snapshots []Snapshot `json:"snapshots"`
}
