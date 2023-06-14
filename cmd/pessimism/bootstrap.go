package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

const (
	extJSON = ".json"
)

type BootSession = models.InvRequestParams

// loadBootStrapFile ... Loads the bootstrap file
func loadBootStrapFile(path string) ([]BootSession, error) {
	if !strings.HasSuffix(path, extJSON) {
		return nil, fmt.Errorf("invalid bootstrap file format; expected %s", extJSON)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data := []BootSession{}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// bootStrap ... Bootstraps the application by starting the invariant sessions
func bootStrap(ctx context.Context, svc service.Service, sessions []BootSession) error {
	logger := logging.WithContext(ctx)

	for _, session := range sessions {
		sUUID, err := svc.RunInvariantSession(session)
		if err != nil {
			return err
		}

		logger.Info("invariant session started",
			zap.String(core.SUUIDKey, sUUID.String()))
	}
	return nil
}
