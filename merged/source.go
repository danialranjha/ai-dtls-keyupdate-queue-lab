// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT
// Exact source function bodies; see source.json for harness boundaries.
package merged

import (
	"context"
	"queue-lab/handshake"
	"slices"
)

func (p *postHandshake) startQueuedPostHandshake(ctx context.Context, conn Conn) error {
	for i := 0; i < len(p.queue); {
		command := p.queue[i]
		if err, canceled := canceledPostHandshakeCommand(command); canceled {
			p.queue = slices.Delete(p.queue, i, i+1)
			command.Completion.complete(err)

			continue
		}
		if !p.commandEligible(command.Kind) {
			// Keep a required response queued until the outstanding update is ACKed.
			// Application records may still use the current keys.
			// https://www.rfc-editor.org/rfc/rfc9147.html#section-8
			i++

			continue
		}
		p.queue = slices.Delete(p.queue, i, i+1)

		err := p.startPostHandshakeCommand(ctx, conn, command)
		if err != nil {
			command.Completion.complete(err)

			return err
		}
	}

	return nil
}

func (p *postHandshake) commandEligible(kind postHandshakeCommandKind) bool {
	var category postHandshakeCategory
	switch kind {
	case commandSendKeyUpdate:
		category = postHandshakeKeyUpdate
	case commandSendNewConnectionID:
		category = postHandshakeNewConnectionID
	case commandSendRequestConnectionID:
		return true
	default:
		// tickets may overlap, and invalid commands must reach validation.
		return true
	}
	for id := range p.flights {
		if id.Category == category {
			return false
		}
	}

	return true
}

func canceledPostHandshakeCommand(command postHandshakeCommand) (error, bool) {
	if command.Canceled == nil {
		return nil, false
	}
	select {
	case <-command.Canceled:
		return context.Canceled, true
	default:
		return nil, false
	}
}

func (p *postHandshake) queueRequiredKeyUpdateResponse(request handshake.KeyUpdateRequest) {
	if request != handshake.KeyUpdateRequested {
		return
	}
	for _, queued := range p.queue {
		if queued.Kind == commandSendKeyUpdate && queued.KeyUpdate.Required {
			return
		}
	}
	command := postHandshakeCommand{Kind: commandSendKeyUpdate, KeyUpdate: keyUpdateCommand{Request: handshake.KeyUpdateNotRequested, Required: true}}
	insertAt := len(p.queue)
	for i, queued := range p.queue {
		if queued.Kind == commandSendApplicationData {
			insertAt = i

			break
		}
	}
	p.queue = append(p.queue, postHandshakeCommand{})
	copy(p.queue[insertAt+1:], p.queue[insertAt:])
	p.queue[insertAt] = command
}
