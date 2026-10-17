// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT
// Exact source function bodies; see source.json for harness boundaries.
package coalesce_only

import (
	"context"
	"queue-lab/handshake"
	"slices"
)

func (p *postHandshake) startQueuedPostHandshake(ctx context.Context, conn Conn) error {
	applicationBlocked := false
	for i := 0; i < len(p.queue); {
		command := p.queue[i]
		if err, canceled := canceledPostHandshakeCommand(command); canceled {
			p.queue = slices.Delete(p.queue, i, i+1)
			command.Completion.complete(err)

			continue
		}
		if !p.commandEligible(command.Kind, applicationBlocked) {
			applicationBlocked = applicationBlocked || command.KeyUpdate.Required
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

func (p *postHandshake) commandEligible(kind postHandshakeCommandKind, applicationBlocked bool) bool {
	var category postHandshakeCategory
	switch kind {
	case commandSendKeyUpdate:
		category = postHandshakeKeyUpdate
	case commandSendNewConnectionID:
		category = postHandshakeNewConnectionID
	case commandSendRequestConnectionID:
		return true
	case commandSendApplicationData:
		return !applicationBlocked
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
