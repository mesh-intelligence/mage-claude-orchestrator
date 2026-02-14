// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"fmt"
	"os"
	"regexp"
)

// versionConstRe matches a Go const declaration like:
//
//	const Version = "v0.20260212.0"
//
// It captures the quoted value.
var versionConstRe = regexp.MustCompile(`(?m)^const\s+Version\s*=\s*"([^"]*)"`)

// readVersionConst reads the Version constant from a Go source file.
// Returns "" if the file does not exist or has no Version constant.
func readVersionConst(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	m := versionConstRe.FindSubmatch(data)
	if m == nil {
		return ""
	}
	return string(m[1])
}

// writeVersionConst updates the Version constant in a Go source file.
// The file must already exist and contain a `const Version = "..."` line.
func writeVersionConst(filePath, version string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading version file: %w", err)
	}

	if !versionConstRe.Match(data) {
		return fmt.Errorf("no Version constant found in %s", filePath)
	}

	updated := versionConstRe.ReplaceAll(data, []byte(fmt.Sprintf(`const Version = "%s"`, version)))
	if err := os.WriteFile(filePath, updated, 0o644); err != nil {
		return fmt.Errorf("writing version file: %w", err)
	}
	return nil
}
