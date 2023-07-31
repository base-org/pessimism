package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// TODO: add timestamp to the message
const (
	CodeBlockFmt = "```%s```"

	// slackMsgFmt ... Slack message format
	SlackMsgFmt = `
	⚠️🚨 Pessimism Alert: %s 🚨⚠️

	_Heuristic activation conditions met_

	_Network:_ %s
	_Session UUID:_ %s

	*Assessment Content:* 
	%s	
	`
)

// Interpolator ... Interface for interpolating messages
type Interpolator interface {
	InterpolateSlackMessage(sUUID core.SUUID, message string) string
}

// interpolator ... Interpolator implementation
type interpolator struct{}

// NewInterpolator ... Initializer
func NewInterpolator() Interpolator {
	return &interpolator{}
}

// InterpolateSlackMessage ... Interpolates a slack message with the given heuristic session UUID and message
func (*interpolator) InterpolateSlackMessage(sUUID core.SUUID, message string) string {
	return fmt.Sprintf(SlackMsgFmt,
		sUUID.PID.HeuristicType().String(),
		sUUID.PID.Network(),
		sUUID.String(),
		fmt.Sprintf(CodeBlockFmt, message))
}
