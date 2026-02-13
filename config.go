// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package orchestrator

// Config holds project-specific settings for the orchestrator.
// Consuming repos create a Config and pass it to New().
type Config struct {
	// ModulePath is the Go module path (e.g., "github.com/mesh-intelligence/crumbs").
	ModulePath string

	// BinaryName is the name of the compiled binary (e.g., "cupboard").
	BinaryName string

	// BinaryDir is the output directory for compiled binaries (default "bin").
	BinaryDir string

	// MainPackage is the path to the main.go entry point
	// (e.g., "cmd/cupboard/main.go").
	MainPackage string

	// GoSourceDirs lists directories containing Go source files
	// (e.g., ["cmd/", "pkg/", "internal/", "tests/"]).
	GoSourceDirs []string

	// VersionFile is the path to the version file
	// (e.g., "pkg/crumbs/version.go").
	VersionFile string

	// GenPrefix is the prefix for generation branch names (default "generation-").
	GenPrefix string

	// BeadsDir is the beads database directory (default ".beads/").
	BeadsDir string

	// CobblerDir is the cobbler scratch directory (default ".cobbler/").
	CobblerDir string

	// MagefilesDir is the directory skipped when deleting Go files
	// (default "magefiles").
	MagefilesDir string

	// SecretsDir is the directory containing token files (default ".secrets").
	SecretsDir string

	// DefaultTokenFile is the default credential filename (default "claude.json").
	DefaultTokenFile string

	// SpecGlobs maps a label to a glob pattern for word-count stats.
	SpecGlobs map[string]string

	// SeedFiles maps relative file paths to content strings.
	// Each file is created during generator:start and generator:reset
	// after deleting Go source code. The content strings are Go
	// text/template templates executed with SeedData.
	SeedFiles map[string]string

	// MeasurePrompt overrides the default measure prompt template.
	// If empty, the embedded default is used.
	MeasurePrompt string

	// StitchPrompt overrides the default stitch prompt template.
	// If empty, the embedded default is used.
	StitchPrompt string
}

// SeedData is the template data passed to SeedFiles templates.
type SeedData struct {
	Version    string
	ModulePath string
}

func (c *Config) applyDefaults() {
	if c.BinaryDir == "" {
		c.BinaryDir = "bin"
	}
	if c.GenPrefix == "" {
		c.GenPrefix = "generation-"
	}
	if c.BeadsDir == "" {
		c.BeadsDir = ".beads/"
	}
	if c.CobblerDir == "" {
		c.CobblerDir = ".cobbler/"
	}
	if c.MagefilesDir == "" {
		c.MagefilesDir = "magefiles"
	}
	if c.SecretsDir == "" {
		c.SecretsDir = ".secrets"
	}
	if c.DefaultTokenFile == "" {
		c.DefaultTokenFile = "claude.json"
	}
}
