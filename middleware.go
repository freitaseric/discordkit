package discordkit

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"runtime/debug"
	"time"
)

// PanicError describes a recovered panic. Stack is only for server-side logging.
// Error intentionally excludes the panic value, which might contain secrets.
type PanicError struct {
	Value any
	Stack []byte
}

func (e *PanicError) Error() string { return "discordkit: handler panicked" }

// Unwrap exposes a panic error to errors.Is and errors.As.
func (e *PanicError) Unwrap() error { v, _ := e.Value.(error); return v }

// Recovery converts handler panics to PanicError; it never replies to a user.
// Router applies Recovery automatically, outside all configured middleware.
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) (err error) {
			defer func() {
				if v := recover(); v != nil {
					err = &PanicError{Value: v, Stack: debug.Stack()}
				}
			}()
			return next(c)
		}
	}
}

// Logging logs latency and errors without interaction tokens, content or options.
func Logging(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next Handler) Handler {
		return func(c *Context) (err error) {
			start := time.Now()
			defer func() {
				logger.InfoContext(c.Context(), "discordkit interaction", "duration", time.Since(start), "error", err)
			}()
			return next(c)
		}
	}
}

// RequireGuild rejects interactions outside a guild.
func RequireGuild() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			if !c.IsGuild() {
				return fmt.Errorf("%w: guild required", ErrInvalidInteraction)
			}
			return next(c)
		}
	}
}

// RequirePermissions checks the invoking member's interaction permissions.
// Administrator bypasses the bit check. This does not check the bot's permissions.
func RequirePermissions(permissions int64) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			m := c.Member()
			if m == nil || (m.Permissions&discordgo.PermissionAdministrator == 0 && m.Permissions&permissions != permissions) {
				return fmt.Errorf("%w: member permissions required", ErrInvalidInteraction)
			}
			return next(c)
		}
	}
}
