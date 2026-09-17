package discordkit

import "errors"

// Sentinel errors identify failures that callers can handle with errors.Is.
var (
	ErrRouteNotFound      = errors.New("discordkit: route not found")
	ErrRouteConflict      = errors.New("discordkit: route conflict")
	ErrInvalidCustomID    = errors.New("discordkit: invalid custom ID")
	ErrAlreadyResponded   = errors.New("discordkit: interaction already acknowledged")
	ErrNotResponded       = errors.New("discordkit: interaction has not been acknowledged")
	ErrInvalidInteraction = errors.New("discordkit: operation is invalid for this interaction")
	ErrInvalidComponent   = errors.New("discordkit: invalid component")
	ErrInvalidMessage     = errors.New("discordkit: invalid message")
	ErrMissingOption      = errors.New("discordkit: option missing or invalid")
	ErrInvalidCommand     = errors.New("discordkit: invalid command")
)
