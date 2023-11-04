package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TODO: add timestamp to the message
// Slack Formatting
const (
	CodeBlockFmt = "```%s```"

	// slackMsgFmt ... Slack message format
	SlackMsgFmt = `
	*%s %s*

	Network: %s
	Severity: %s
	Session UUID: %s

	*Assessment Content:* 
	%s	

	*Message:*
	%s
	`
)

const (
	PagerDutyMsgFmt = `
	Heuristic Triggered: %s
	Network: %s
	Assessment: 
	%s
	`
)

// Interpolator ... Interface for interpolating messages
type Interpolator interface {
	InterpolateSlackMessage(sev core.Severity, sUUID core.UUID, content string, msg string) string
	InterpolatePagerDutyMessage(sUUID core.UUID, message string) string
}

// interpolator ... Interpolator implementation
type interpolator struct{}

// NewInterpolator ... Initializer
func NewInterpolator() Interpolator {
	return &interpolator{}
}

// InterpolateSlackMessage ... Interpolates a slack message with the given heuristic session UUID and message
func (*interpolator) InterpolateSlackMessage(sev core.Severity, sUUID core.UUID, content string, msg string) string {
	return fmt.Sprintf(SlackMsgFmt,
		sev.Symbol(),
		// sUUID.PID.HeuristicType().String(),
		// sUUID.PID.Network(),
		cases.Title(language.English).String(sev.String()),
		sUUID.String(),
		fmt.Sprintf(CodeBlockFmt, content),
		msg)
}

// InterpolatePagerDutyMessage ... Interpolates a pagerduty message with the given heuristic session UUID and message
func (*interpolator) InterpolatePagerDutyMessage(sUUID core.UUID, message string) string {
	return fmt.Sprintf(PagerDutyMsgFmt,
		// sUUID.PID.HeuristicType().String(),
		// sUUID.PID.Network(),
		message)
}
