//go:build e2e

// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mesh-intelligence/cobbler-scaffold/pkg/orchestrator"
)

// BenchmarkRel01_UC008_MeasureTimingByLimit runs measure with limits 1 through 5
// and reports wall-clock time and issue count for each.
//
//	go test -tags e2e -bench BenchmarkRel01_UC008_MeasureTimingByLimit -benchtime 1x -timeout 0 ./tests/rel01.0/...
func BenchmarkRel01_UC008_MeasureTimingByLimit(b *testing.B) {
	dir := setupRepo(b)
	setupClaude(b, dir)

	if err := runMage(b, dir, "reset"); err != nil {
		b.Fatalf("reset: %v", err)
	}
	if err := runMage(b, dir, "init"); err != nil {
		b.Fatalf("init: %v", err)
	}

	for limit := 1; limit <= 5; limit++ {
		b.Run(fmt.Sprintf("limit_%d", limit), func(b *testing.B) {
			b.StopTimer()
			if err := runMage(b, dir, "beads:reset"); err != nil {
				b.Fatalf("beads:reset: %v", err)
			}
			if err := runMage(b, dir, "init"); err != nil {
				b.Fatalf("init: %v", err)
			}

			writeConfigOverride(b, dir, func(cfg *orchestrator.Config) {
				cfg.Cobbler.MaxMeasureIssues = limit
			})
			b.StartTimer()

			for range b.N {
				if err := runMage(b, dir, "cobbler:measure"); err != nil {
					b.Fatalf("cobbler:measure (limit=%d): %v", limit, err)
				}
			}

			b.StopTimer()
			n := countReadyIssues(b, dir)
			b.ReportMetric(float64(n), "issues")
		})
	}
}

// BenchmarkRel01_UC008_StitchTimingByCycle runs alternating measure/stitch
// cycles and reports wall-clock time for each phase.
//
//	go test -tags e2e -bench BenchmarkRel01_UC008_StitchTimingByCycle -benchtime 1x -timeout 0 ./tests/rel01.0/...
func BenchmarkRel01_UC008_StitchTimingByCycle(b *testing.B) {
	dir := setupRepo(b)
	setupClaude(b, dir)

	const cycles = 5

	writeConfigOverride(b, dir, func(cfg *orchestrator.Config) {
		cfg.Cobbler.MaxMeasureIssues = 1
		cfg.Cobbler.MaxStitchIssuesPerCycle = 1
		cfg.Claude.MaxTimeSec = 600
	})

	if err := runMage(b, dir, "reset"); err != nil {
		b.Fatalf("reset: %v", err)
	}
	if err := runMage(b, dir, "generator:start"); err != nil {
		b.Fatalf("generator:start: %v", err)
	}

	for range b.N {
		for i := 1; i <= cycles; i++ {
			b.Run(fmt.Sprintf("cycle_%d", i), func(b *testing.B) {
				mStart := time.Now()
				if err := runMage(b, dir, "cobbler:measure"); err != nil {
					b.Fatalf("cycle %d measure: %v", i, err)
				}
				mElapsed := time.Since(mStart).Round(time.Second)

				n := countReadyIssues(b, dir)
				if n == 0 {
					b.Fatalf("cycle %d: expected at least 1 ready issue after measure, got 0", i)
				}
				b.ReportMetric(float64(mElapsed.Seconds()), "measure_sec")

				sStart := time.Now()
				if err := runMage(b, dir, "cobbler:stitch"); err != nil {
					b.Fatalf("cycle %d stitch: %v", i, err)
				}
				sElapsed := time.Since(sStart).Round(time.Second)
				b.ReportMetric(float64(sElapsed.Seconds()), "stitch_sec")
				b.ReportMetric(float64(n), "issues")
			})
		}
	}
}
