package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// loadBootStrapFile ... Loads the bootstrap file
func loadBootStrapFile(path string) ([]models.InvRequestParams, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data := []models.InvRequestParams{}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// bootStrap ... Bootstraps the application by starting the invariant sessions
func bootStrap(ctx context.Context, svc service.Service, params []models.InvRequestParams) error {
	logger := logging.WithContext(ctx)

	for _, param := range params {
		sUUID, err := svc.RunInvariantSession(param)
		if err != nil {
			return err
		}

		logger.Info("invariant session started",
			zap.String(core.SUUIDKey, sUUID.String()))
	}
	return nil
}
