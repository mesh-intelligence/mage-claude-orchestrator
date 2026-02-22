# eng05: Stitch Benchmark Analysis — Why Claude Stalls in Later Cycles

## Summary

A 5-cycle benchmark (`TestStitch_TimingByCycle`) ran alternating measure/stitch
passes with `max_time_sec=600` (10 minutes) per Claude invocation. Only cycle 1
completed stitch within timeout. Cycles 2–5 all hit the 10-minute wall.

| Cycle | Measure | Stitch  | Outcome           |
|-------|---------|---------|-------------------|
| 1     | 1m22s   | 4m14s   | completed         |
| 2     | 1m35s   | 10m3s   | timeout           |
| 3     | 2m7s    | 10m3s   | timeout           |
| 4     | 1m57s   | 10m3s   | timeout           |
| 5     | 2m7s    | 10m2s   | timeout           |

Total wall clock: 53m33s. Total stitch time: 44m25s, of which only 4m14s
produced a merged commit.

## Root Causes

### 1. Measure generates tasks that grow in scope with each cycle

Each measure call sees the full project context plus all existing issues. As
the project accumulates code from prior cycles, the next task the planner
proposes is contextually aware of more surface area. In practice, cycle 1
produced a modest `pkg/testutils` (3 files, 839 lines). Cycle 2 proposed
`cmd/head`, cycle 4 `cmd/wc`, cycle 5 `cmd/cat` — each requiring cross-platform
output-format parity with GNU coreutils.

The planning constitution says "target 300–700 lines of production code,
touching no more than 5 files" (P2), but measure does not enforce this at
import time. The stitch agent receives descriptions that were sized by the
planner's estimate, but real implementation complexity exceeds the estimate
because:

- Format-matching requirements (e.g., align `wc` column widths with GNU wc)
  trigger iterative test-fix-test loops the planner cannot predict.
- Each utility needs its own test suite, CLI wiring, and integration with
  existing project patterns, which pushes actual scope well beyond 700 lines.

### 2. Stitch agent reads too much before writing

The stitch prompt injects the entire project context (docs, specs, source code)
via `buildProjectContext`. As the codebase grows cycle over cycle, this context
payload grows. The stitch agent is told "Review the files listed in Required
Reading by finding them in the PROJECT CONTEXT," but Claude's tendency is to
also read files via tools — despite the prompt saying not to — to verify that
the inline context is current. This manifests as:

- Turns 1–5 (2–3 minutes): Claude reads source files already provided inline.
- Turns 6–10: Claude reads test files, Makefiles, existing utilities.
- Only after minute 4–5 does actual file creation begin.

In cycle 5 (`cmd/cat`), the stream-json log showed Claude spending 9+ minutes
on setup and reading before writing a single file, then hitting the 32K output
token limit.

### 3. Output token ceiling creates a hard wall

Claude's 32K output token limit is a constraint the orchestrator cannot
control. When a stitch task requires many turns (reading, planning, writing
multiple files, running tests, fixing test failures), the cumulative output
approaches this ceiling. Cycle 5 hit it explicitly ("API Error: Claude's
response exceeded the 32000 output token maximum"). Once Claude runs out of
output tokens, the task fails regardless of time remaining.

### 4. Test-fix-test loops are unbounded

The execution constitution (E5) says "Every item in Acceptance Criteria must be
verified. Run tests if the criteria require it." Combined with acceptance
criteria like "all tests pass" and "output matches GNU `wc`", Claude enters
a loop:

1. Write implementation.
2. Run tests.
3. Tests fail on formatting edge case.
4. Fix formatting.
5. Go to 2.

In cycle 4, the stream-json showed 63 turns, mostly iterating on column-width
alignment for `wc` output. This loop has no termination bound within the
stitch prompt.

## Recommendations

### A. Reduce task scope at the measure level

The planning constitution P2 says 300–700 lines. For an automated pipeline
with a 10-minute timeout, this is too large. Recommend:

- **Lower `estimated_lines_min`/`estimated_lines_max`** from 250–350 to
  100–200 in the config. Measure already templates these as `{{.LinesMin}}`
  and `{{.LinesMax}}`.
- **Add explicit constraint**: "Each task must be completable by Claude in
  under 5 minutes. Prefer creating a struct and its methods as one task, and
  tests as a second task."
- **Split test tasks from implementation tasks**: A task that says "implement X
  and test X" is two tasks. The stitch agent should implement first; a
  follow-up task tests it.

### B. Cap stitch agent turns and output

Add a `max_turns` constraint to the stitch prompt or the Claude CLI arguments:

- `--max-turns 20` would force Claude to finish within 20 tool-use rounds.
  If it cannot complete in 20 turns, the task is too large.
- Consider reducing `max_time_sec` to 300 (5 minutes) and treating timeout
  as a signal that the task needs splitting, not retrying.

### C. Suppress redundant file reads in the stitch prompt

The stitch prompt already says "Do NOT read files already provided in PROJECT
CONTEXT." Reinforce this:

- Add a louder constraint: "You MUST NOT use Read, Bash cat, or any file
  reading tool on files listed in source_code above. They are already inline.
  Doing so wastes your limited output tokens."
- Consider *not* providing full source code in the stitch prompt for smaller
  tasks. If the task only touches 2–3 files, include only those files plus
  interfaces they depend on.

### D. Reduce project context size for stitch

`buildProjectContext` loads *all* docs, specs, constitutions, engineering
guidelines, and source files. For a task like "implement `cmd/head`", most of
this context is irrelevant. Options:

- **Selective context**: Use the task's `required_reading` list to include only
  the files the task actually needs, plus a summary of project structure.
- **Context budget**: Set a maximum context size (e.g., 50K tokens). If the
  full context exceeds this, truncate source files that are not in
  `required_reading`.

### E. Add a format-matching escape valve

Tasks involving output-format parity with existing tools (GNU coreutils) are
inherently unbounded. The measure prompt should:

- Avoid proposing tasks that require exact output-format matching with
  external tools.
- If format matching is needed, provide the expected output inline in the task
  description rather than expecting Claude to discover it by running the
  external tool.

## Metrics

From the benchmark run:

- **Cost**: Cycle 1 stitch: $1.94 (completed). Cycles 2–5 stitch: each
  consumed full timeout budget with no merge.
- **Effective throughput**: 1 task completed per 53 minutes of wall-clock time.
  Target should be 1 task per 5–10 minutes.
- **Token efficiency**: Cycle 1 used 21 turns. Cycle 4 used 63 turns (3x),
  mostly on formatting iteration.

## Next Steps

1. Reduce `estimated_lines_min`/`estimated_lines_max` to 100–200.
2. Add `--max-turns 25` to Claude CLI args.
3. Split "implement + test" tasks in the planning constitution.
4. Run the benchmark again and compare cycle completion rates.
