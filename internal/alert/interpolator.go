package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// TODO: add timestamp to the message
const (
	slackMsgFmt = `
	⚠️🚨 Pessimism Alert: %s Invariant Invalidation 🚨⚠️

	_Invariant invalidation conditions met_

	*Network:* %s
	*Session UUID:* %s

	*Assessment Content:* 
	%s
	
	_Remember to check the logs for more information and to take action if necessary. You can't always be optimistic!_
	¯\_(ツ)_/¯
	`
)

type Interpolator interface {
	InterpolateSlackMessage(sUUID core.InvSessionUUID, message string) string
}

type interpolator struct {
}

func NewInterpolator() Interpolator {
	return &interpolator{}
}

func (i *interpolator) InterpolateSlackMessage(sUUID core.InvSessionUUID, message string) string {
	return fmt.Sprintf(slackMsgFmt,
		fmt.Sprintf("`%s`", sUUID.PID.InvType().String()),
		sUUID.PID.Network(),
		sUUID.String(),
		message)
}
