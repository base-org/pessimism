package alert

import (
	"time"

	"github.com/base-org/pessimism/internal/core"
)

// CoolDownHandler ... Interface for the cool down handler
type CoolDownHandler interface {
	Add(id core.UUID, coolDownTime time.Duration)
	Update()
	IsCoolDown(id core.UUID) bool
}

// coolDownHandler ... Implementation of CoolDownHandler
type coolDownHandler struct {
	sessions map[core.UUID]time.Time
}

// NewCoolDownHandler ... Initializer
func NewCoolDownHandler() CoolDownHandler {
	return &coolDownHandler{
		sessions: make(map[core.UUID]time.Time),
	}
}

// Add ... Adds a session to the cool down handler
func (cdh *coolDownHandler) Add(id core.UUID, coolDownTime time.Duration) {
	cdh.sessions[id] = time.Now().Add(coolDownTime)
}

// Update ... Updates the cool down handler
func (cdh *coolDownHandler) Update() {
	for id, t := range cdh.sessions {
		if t.Before(time.Now()) {
			delete(cdh.sessions, id)
		}
	}
}

// IsCoolDown ... Checks if the session is in cool down
func (cdh *coolDownHandler) IsCoolDown(id core.UUID) bool {
	if t, ok := cdh.sessions[id]; ok {
		return t.After(time.Now())
	}

	return false
}
