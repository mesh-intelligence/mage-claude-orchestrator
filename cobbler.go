// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CobblerConfig holds options shared by measure and stitch targets.
type CobblerConfig struct {
	SilenceAgent     bool
	MaxIssues        int
	UserPrompt       string
	GenerationBranch string
	TokenFile        string
}

// logConfig prints the resolved configuration for debugging.
func (c *CobblerConfig) logConfig(target string) {
	logf("%s config: silenceAgent=%v maxIssues=%d tokenFile=%s generationBranch=%q",
		target, c.SilenceAgent, c.MaxIssues, c.TokenFile, c.GenerationBranch)
	if c.UserPrompt != "" {
		logf("%s config: userPrompt=%q", target, c.UserPrompt)
	}
}

// registerCobblerFlags adds the shared flags to fs.
func (o *Orchestrator) registerCobblerFlags(fs *flag.FlagSet, cfg *CobblerConfig) {
	fs.BoolVar(&cfg.SilenceAgent, flagSilenceAgent, true, "suppress Claude output")
	fs.IntVar(&cfg.MaxIssues, flagMaxIssues, 10, "max issues to process")
	fs.StringVar(&cfg.UserPrompt, flagUserPrompt, "", "user prompt text")
	fs.StringVar(&cfg.GenerationBranch, flagGenerationBranch, "", "generation branch to work on")
	fs.StringVar(&cfg.TokenFile, flagTokenFile, o.cfg.DefaultTokenFile, "token file name in .secrets/")
}

// resolveCobblerBranch sets cfg.GenerationBranch from the first positional arg
// if the flag was not provided.
func resolveCobblerBranch(cfg *CobblerConfig, fs *flag.FlagSet) {
	if cfg.GenerationBranch == "" && fs.NArg() > 0 {
		cfg.GenerationBranch = fs.Arg(0)
	}
}

// ClaudeResult holds token usage from a Claude invocation.
type ClaudeResult struct {
	InputTokens  int
	OutputTokens int
}

// LocSnapshot holds a point-in-time LOC count.
type LocSnapshot struct {
	Production int `json:"production"`
	Test       int `json:"test"`
}

// captureLOC returns the current Go LOC counts. Errors are swallowed
// because stats collection is best-effort.
func (o *Orchestrator) captureLOC() LocSnapshot {
	rec, err := o.CollectStats()
	if err != nil {
		logf("captureLOC: collectStats error: %v", err)
		return LocSnapshot{}
	}
	return LocSnapshot{Production: rec.GoProdLOC, Test: rec.GoTestLOC}
}

// InvocationRecord is the JSON blob recorded as a beads comment after
// every Claude invocation.
type InvocationRecord struct {
	Caller    string       `json:"caller"`
	StartedAt string      `json:"started_at"`
	DurationS int         `json:"duration_s"`
	Tokens    claudeTokens `json:"tokens"`
	LOCBefore LocSnapshot  `json:"loc_before"`
	LOCAfter  LocSnapshot  `json:"loc_after"`
	Diff      diffRecord   `json:"diff"`
}

type claudeTokens struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type diffRecord struct {
	Files      int `json:"files"`
	Insertions int `json:"insertions"`
	Deletions  int `json:"deletions"`
}

// recordInvocation serializes an InvocationRecord to JSON and adds it
// as a beads comment on the given issue.
func recordInvocation(issueID string, rec InvocationRecord) {
	data, err := json.Marshal(rec)
	if err != nil {
		logf("recordInvocation: marshal error: %v", err)
		return
	}
	if err := bdCommentAdd(issueID, string(data)); err != nil {
		logf("recordInvocation: bd comment error for %s: %v", issueID, err)
	}
}

// parseClaudeTokens extracts token usage from Claude's stream-json
// output. The final JSON line has "type":"result" with a "usage" object
// containing "input_tokens" and "output_tokens".
func parseClaudeTokens(output []byte) ClaudeResult {
	lines := bytes.Split(bytes.TrimSpace(output), []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		var msg struct {
			Type  string `json:"type"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(lines[i], &msg); err != nil {
			continue
		}
		if msg.Type == "result" {
			return ClaudeResult{
				InputTokens:  msg.Usage.InputTokens,
				OutputTokens: msg.Usage.OutputTokens,
			}
		}
	}
	return ClaudeResult{}
}

// runClaude executes Claude with the given prompt and returns token usage.
func (o *Orchestrator) runClaude(prompt, dir string, silence bool) (ClaudeResult, error) {
	logf("runClaude: promptLen=%d dir=%q silence=%v", len(prompt), dir, silence)
	logf("runClaude: exec %s %v", binClaude, claudeArgs)
	cmd := exec.Command(binClaude, claudeArgs...)
	cmd.Stdin = strings.NewReader(prompt)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdoutBuf bytes.Buffer
	if silence {
		cmd.Stdout = &stdoutBuf
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = os.Stderr
	}

	start := time.Now()
	err := cmd.Run()
	result := parseClaudeTokens(stdoutBuf.Bytes())
	logf("runClaude: finished in %s tokens(in=%d out=%d) (err=%v)",
		time.Since(start).Round(time.Second), result.InputTokens, result.OutputTokens, err)
	return result, err
}

// worktreeBasePath returns the directory used for stitch worktrees.
func worktreeBasePath() string {
	repoRoot, _ := os.Getwd()
	return filepath.Join(os.TempDir(), filepath.Base(repoRoot)+"-worktrees")
}

// CobblerReset removes the cobbler scratch directory.
func (o *Orchestrator) CobblerReset() error {
	logf("cobblerReset: removing %s", o.cfg.CobblerDir)
	os.RemoveAll(o.cfg.CobblerDir)
	logf("cobblerReset: done")
	return nil
}

// beadsCommit syncs beads state and commits the beads directory.
func (o *Orchestrator) beadsCommit(msg string) {
	logf("beadsCommit: %s", msg)
	if err := bdSync(); err != nil {
		logf("beadsCommit: bdSync warning: %v", err)
	}
	if err := gitStageDir(o.cfg.BeadsDir); err != nil {
		logf("beadsCommit: gitStageDir warning: %v", err)
	}
	if err := gitCommitAllowEmpty(msg); err != nil {
		logf("beadsCommit: gitCommit warning: %v", err)
	}
	logf("beadsCommit: done")
}
