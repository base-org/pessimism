package alert

import (
	"time"

	"github.com/base-org/pessimism/internal/core"
)

// CoolDownHandler ... Interface for the cool down handler
type CoolDownHandler interface {
	Add(suuid core.SUUID, coolDownTime time.Duration)
	Update()
	IsCoolDown(suuid core.SUUID) bool
}

// coolDownHandler ... Implementation of CoolDownHandler
type coolDownHandler struct {
	sessions map[core.SUUID]time.Time
}

// NewCoolDownHandler ... Initializer
func NewCoolDownHandler() CoolDownHandler {
	return &coolDownHandler{
		sessions: make(map[core.SUUID]time.Time),
	}
}

// Add ... Adds a session to the cool down handler
func (cdh *coolDownHandler) Add(sUUID core.SUUID, coolDownTime time.Duration) {
	cdh.sessions[sUUID] = time.Now().Add(coolDownTime)
}

// Update ... Updates the cool down handler
func (cdh *coolDownHandler) Update() {
	for sUUID, t := range cdh.sessions {
		if t.Before(time.Now()) {
			delete(cdh.sessions, sUUID)
		}
	}
}

// IsCoolDown ... Checks if the session is in cool down
func (cdh *coolDownHandler) IsCoolDown(sUUID core.SUUID) bool {
	if t, ok := cdh.sessions[sUUID]; ok {
		return t.After(time.Now())
	}

	return false
}
