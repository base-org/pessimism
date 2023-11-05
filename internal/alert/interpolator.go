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

type Interpolator struct{}

func NewInterpolator() *Interpolator {
	return &Interpolator{}
}

func (*Interpolator) SlackMessage(a core.Alert, msg string) string {
	return fmt.Sprintf(SlackMsgFmt,
		a.Sev.Symbol(),
		a.HT.String(),
		a.Net.String(),
		cases.Title(language.English).String(a.Sev.String()),
		a.HeuristicID.String(),
		fmt.Sprintf(CodeBlockFmt, a.Content),
		msg)
}

func (*Interpolator) PagerDutyMessage(a core.Alert) string {
	return fmt.Sprintf(PagerDutyMsgFmt,
		a.HT.String(),
		a.Net.String(),
		a.Content)
}
