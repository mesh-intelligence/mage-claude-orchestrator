<!-- Copyright (c) 2026 Petar Djukic. All rights reserved. SPDX-License-Identifier: MIT -->

# Prompt Template Conventions

## Introduction

We use Go text/template for both the measure and stitch prompts. Default templates are embedded in the binary via `//go:embed`. Consuming projects can override them through Config fields. This guideline documents the template data contracts, customization patterns, and conventions for writing effective prompts.

## Embedded Defaults

The orchestrator embeds two templates from the `pkg/orchestrator/prompts/` directory.

Table 1 Embedded Prompt Templates

| Template | File                                   | Data Type         | Purpose                              |
|----------|----------------------------------------|-------------------|--------------------------------------|
| Measure  | pkg/orchestrator/prompts/measure.tmpl  | MeasurePromptData | Propose new tasks from project state |
| Stitch   | pkg/orchestrator/prompts/stitch.tmpl   | StitchPromptData  | Execute a single task                |

We embed these files using `//go:embed` directives in pkg/orchestrator/measure.go and pkg/orchestrator/stitch.go. The embedded strings serve as defaults when Config.MeasurePrompt or Config.StitchPrompt is empty.

## Template Data Contracts

### MeasurePromptData

Table 2 MeasurePromptData Fields

| Field | Type | Source |
|-------|------|--------|
| ExistingIssues | string (JSON) | bd list output |
| Limit | int | Config.MaxMeasureIssues |
| OutputPath | string | Computed file path in Config.CobblerDir |
| UserInput | string | Config.UserPrompt |
| LinesMin | int | Config.EstimatedLinesMin |
| LinesMax | int | Config.EstimatedLinesMax |
| ProjectRules | string | Concatenated .claude/rules/*.md files |

The measure template receives a JSON string of existing issues. We render it inline in the prompt so Claude sees the full issue tracker state. The Limit field tells Claude how many tasks to propose. The OutputPath is where Claude writes its JSON response. The ProjectRules field contains all project rules from `.claude/rules/` so Claude understands project conventions.

### StitchPromptData

Table 3 StitchPromptData Fields

| Field | Type | Source |
|-------|------|--------|
| Title | string | Task title from beads |
| ID | string | Task ID from beads |
| IssueType | string | Task type from beads (default "task") |
| Description | string | Task description from beads |
| ProjectRules | string | Concatenated .claude/rules/*.md files |

The stitch template receives the task details. Claude uses these to understand what to implement and includes the task ID in commit messages. The ProjectRules field provides project conventions so the stitch agent follows the same style and patterns as human developers.

## Customization

Consuming projects override templates by setting `measure_prompt` or `stitch_prompt` in `configuration.yaml` to a file path. During `LoadConfig`, the orchestrator reads the file and stores its content in the Config field. The file content must be a valid Go text/template that uses the corresponding data type.

```yaml
# configuration.yaml
measure_prompt: "templates/measure.tmpl"
stitch_prompt: "templates/stitch.tmpl"
```

When writing custom templates, we reference the data fields using `{{.FieldName}}` syntax. Conditional sections use `{{- if .UserInput}}...{{- end}}`.

## Conventions for Prompt Authors

We follow these conventions when writing or modifying prompt templates.

We instruct Claude to read project documentation (VISION, ARCHITECTURE, PRDs) before acting. This ensures generated code aligns with the project specifications.

We include structured output instructions. The measure template specifies a JSON format for proposed tasks. The stitch template instructs Claude to commit with the task ID.

We avoid prescribing specific implementation details in the prompt. The specifications (PRDs, use cases) carry the requirements. The prompt points Claude to those documents.

We use the UserInput field for session-specific context that does not belong in the template itself.

## Prompt Structure Convention

We structure all prompt templates using the Structural Trio pattern: Role, Task, and Constraints. This pattern produces more consistent and predictable output from Claude by separating identity from instructions from boundaries.

### Section Layout

Every prompt template uses top-level H1 headings separated by `---` delimiters.

Table 4 Prompt Sections

| Section | Purpose | Required |
|---------|---------|----------|
| ROLE | Who the agent is, what it can see, what it cannot see | yes |
| CONTEXT | Project rules, existing state, task metadata | yes |
| TASK | Numbered chain-of-thought steps the agent must follow in order | yes |
| CONSTRAINTS | Negative instructions: what the agent must NOT do | yes |
| OUTPUT FORMAT | JSON schema, file format, or description structure (measure only) | measure only |
| DESCRIPTION | The task description from the issue tracker (stitch only) | stitch only |
| ADDITIONAL CONTEXT | User-provided session context (measure only, conditional) | no |

### Chain-of-Thought Steps

The TASK section uses numbered steps that force the agent to reason before acting. Each step builds on the previous one. The measure template uses five steps: read, summarize, reason about priorities, propose tasks, write output. The stitch template uses five steps: read required files, plan approach, implement, verify, commit.

Numbered steps prevent the agent from jumping to output without analysis. Step 2 in the measure template ("Summarize project state") forces a think-aloud phase that grounds the subsequent proposals in observed project state rather than assumptions.

### Negative Constraints

The CONSTRAINTS section uses explicit "Do NOT" instructions. These prevent common failure modes.

Table 5 Constraint Categories

| Category | Example | Why |
|----------|---------|-----|
| Tool restrictions | Do NOT use bd commands | Agent should write JSON, not interact with issue tracker |
| Scope boundaries | Do NOT modify files outside the list | Prevents uncontrolled changes |
| Context isolation | Do NOT assume the stitch agent has access to your analysis | Measure must write self-contained descriptions |
| Quality gates | Do NOT leave uncommitted changes | Ensures clean state after stitch |

### Structured Issue Descriptions

The measure template specifies a description schema that every proposed task must follow. This schema aligns with the `.claude/rules/crumb-format.md` rule and ensures that the stitch agent receives uniform, self-contained task descriptions.

Table 6 Description Schema Sections

| Section | Purpose |
|---------|---------|
| Required Reading | Files the stitch agent must read before starting |
| Files to Create/Modify | Explicit scope of changes with create/modify annotation |
| Requirements | Numbered, verb-leading requirements |
| Design Decisions | Patterns and architectural constraints to follow |
| Acceptance Criteria | Checkable pass/fail criteria |

The description schema bridges the measure and stitch phases. Measure produces structured descriptions; stitch consumes them. Because the stitch agent sees only its task description and project rules, the description must contain everything the agent needs to complete the work.

### Delimiter Convention

We use `---` (horizontal rule) between top-level sections. This provides visual separation in the rendered prompt and helps Claude identify section boundaries. Within sections, we use standard markdown formatting (headers, lists, code blocks).
