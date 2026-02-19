// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"strings"
	"testing"
)

func TestMeasurePromptIncludesPlanningConstitution(t *testing.T) {
	o := New(Config{})
	prompt := o.buildMeasurePrompt("", "[]", 5, "/tmp/out.yaml")

	if !strings.Contains(prompt, "## Planning Constitution") {
		t.Error("measure prompt missing '## Planning Constitution' section")
	}
	if !strings.Contains(prompt, "```yaml") {
		t.Error("measure prompt missing YAML code fence for constitution")
	}
	// Check for a key planning constitution article
	if !strings.Contains(prompt, "Release-driven priority") {
		t.Error("measure prompt missing planning constitution content (article P1)")
	}
}

func TestMeasurePromptOmitsVisionAndArchitectureWhenMissing(t *testing.T) {
	o := New(Config{})
	prompt := o.buildMeasurePrompt("", "[]", 5, "/tmp/out.yaml")

	// When VISION.yaml and ARCHITECTURE.yaml don't exist, sections are omitted
	if strings.Contains(prompt, "## Vision") {
		t.Error("measure prompt should not include Vision section when file is missing")
	}
	if strings.Contains(prompt, "## Architecture") {
		t.Error("measure prompt should not include Architecture section when file is missing")
	}
}

func TestStitchPromptIncludesExecutionConstitution(t *testing.T) {
	o := New(Config{})
	task := stitchTask{
		id:          "test-001",
		title:       "Test task",
		issueType:   "task",
		description: "A test description.",
		worktreeDir: "/tmp",
	}

	prompt := o.buildStitchPrompt(task)

	if !strings.Contains(prompt, "## Execution Constitution") {
		t.Error("stitch prompt missing '## Execution Constitution' section")
	}
	if !strings.Contains(prompt, "```yaml") {
		t.Error("stitch prompt missing YAML code fence for constitution")
	}
	// Check for a key execution constitution article
	if !strings.Contains(prompt, "Specification-first") {
		t.Error("stitch prompt missing execution constitution content (article E1)")
	}
}

func TestStitchPromptIncludesTaskContext(t *testing.T) {
	o := New(Config{})
	task := stitchTask{
		id:          "task-123",
		title:       "Implement feature X",
		issueType:   "task",
		description: "Detailed requirements here.",
		worktreeDir: "/tmp",
	}

	prompt := o.buildStitchPrompt(task)

	if !strings.Contains(prompt, "task-123") {
		t.Error("stitch prompt missing task ID")
	}
	if !strings.Contains(prompt, "Implement feature X") {
		t.Error("stitch prompt missing task title")
	}
	if !strings.Contains(prompt, "Detailed requirements here.") {
		t.Error("stitch prompt missing task description")
	}
}
