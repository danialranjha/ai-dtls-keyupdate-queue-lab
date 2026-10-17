// Synthetic harness, not Pion transport or cryptographic implementation.
package original

import (
	"context"
	"queue-lab/fixture"
	"queue-lab/handshake"
)

type postHandshakeCategory uint8

const (
	postHandshakeNewSessionTicket postHandshakeCategory = iota
	postHandshakeKeyUpdate
	postHandshakeNewConnectionID
)

type postHandshakeFlightID struct{ Category postHandshakeCategory }
type postHandshakeCommandKind uint8

const (
	commandSendNewSessionTicket postHandshakeCommandKind = iota
	commandSendKeyUpdate
	commandSendNewConnectionID
	commandSendRequestConnectionID
	commandSendApplicationData
)

type keyUpdateCommand struct {
	Request  handshake.KeyUpdateRequest
	Required bool
}
type completion struct{}

func (c *completion) complete(err error) {}

type postHandshakeCommand struct {
	Kind       postHandshakeCommandKind
	KeyUpdate  keyUpdateCommand
	Canceled   <-chan struct{}
	Completion *completion
	ID         string
}
type Conn struct{}
type postHandshake struct {
	queue   []postHandshakeCommand
	flights map[postHandshakeFlightID]bool
	started []string
}

func (p *postHandshake) startPostHandshakeCommand(ctx context.Context, c Conn, cmd postHandshakeCommand) error {
	label := cmd.ID
	if cmd.Kind == commandSendKeyUpdate {
		label = "required-response"
		if !cmd.KeyUpdate.Required {
			label = "local-update"
		}
		p.flights[postHandshakeFlightID{postHandshakeKeyUpdate}] = true
	}
	p.started = append(p.started, label)
	return nil
}
func (p *postHandshake) snapshot(label string) fixture.Snapshot {
	q := []string{}
	required := 0
	for _, c := range p.queue {
		n := c.ID
		if c.Kind == commandSendKeyUpdate {
			n = "local-update"
			if c.KeyUpdate.Required {
				n = "required-response"
				required++
			}
		}
		q = append(q, n)
	}
	started := append([]string{}, p.started...)
	return fixture.Snapshot{Phase: label, Queue: q, Started: started, RequiredQueued: required, KeyUpdateInFlight: p.flights[postHandshakeFlightID{postHandshakeKeyUpdate}]}
}
func Run(f fixture.Case) fixture.Result {
	p := postHandshake{flights: map[postHandshakeFlightID]bool{}, queue: []postHandshakeCommand{}, started: []string{}}
	if f.InFlight {
		p.flights[postHandshakeFlightID{postHandshakeKeyUpdate}] = true
	}
	if f.OtherFlight {
		p.flights[postHandshakeFlightID{postHandshakeNewConnectionID}] = true
	}
	if f.LocalQueued {
		p.queue = append(p.queue, postHandshakeCommand{Kind: commandSendKeyUpdate, ID: "local-update"})
	}
	c := postHandshakeCommand{Kind: commandSendApplicationData, ID: "ordinary-write"}
	if f.CancelWrite {
		ch := make(chan struct{})
		close(ch)
		c.Canceled = ch
	}
	p.queue = append(p.queue, c)
	for i := 0; i < f.Requests; i++ {
		request := handshake.KeyUpdateRequested
		if f.NotRequested {
			request = handshake.KeyUpdateNotRequested
		}
		p.queueRequiredKeyUpdateResponse(request)
	}
	out := fixture.Result{Case: f.Name, Mode: "original", Snapshots: []fixture.Snapshot{p.snapshot("queued")}}
	if err := p.startQueuedPostHandshake(context.Background(), Conn{}); err != nil {
		panic(err)
	}
	out.Snapshots = append(out.Snapshots, p.snapshot("first-drive"))
	for i := 0; i < f.Acks; i++ {
		delete(p.flights, postHandshakeFlightID{postHandshakeKeyUpdate})
		if err := p.startQueuedPostHandshake(context.Background(), Conn{}); err != nil {
			panic(err)
		}
		out.Snapshots = append(out.Snapshots, p.snapshot("ack-and-drive"))
	}
	return out
}
