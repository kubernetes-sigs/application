// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func PathToOSFile(relativPath string) (*os.File, error) {
	path, err := filepath.Abs(relativPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed generate absolut file path of %s", relativPath))
	}

	manifest, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open file %s", path))
	}

	return manifest, nil
}
