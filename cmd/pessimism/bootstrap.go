package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/base-org/pessimism/internal/app"
)

const (
	extJSON = ".json"
)

// fetchBootSessions ... Loads the bootstrap file
func fetchBootSessions(path string) ([]app.BootSession, error) {
	if !strings.HasSuffix(path, extJSON) {
		return nil, fmt.Errorf("invalid bootstrap file format; expected %s", extJSON)
	}

	file, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	data := []app.BootSession{}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
