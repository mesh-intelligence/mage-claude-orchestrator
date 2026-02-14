<!-- Copyright (c) 2026 Petar Djukic. All rights reserved. SPDX-License-Identifier: MIT -->

# Git Tagging Convention

We use a semver-inspired tagging scheme: `v[REL].[DATE].[REVISION]`.

| Segment  | Description                      | Example  |
|----------|----------------------------------|----------|
| REL      | Release number (see table below) | 0, 1, 2  |
| DATE     | Date in YYYYMMDD format          | 20260213 |
| REVISION | Revision within the same date    | 0, 1, 2  |

Examples: `v0.20260213.0`, `v0.20260213.1`, `v1.20260301.0`.

The REVISION starts at 0 for each new date and increments if multiple tags are created on the same date.

## REL Values

| REL | Purpose                                  | Created by     |
|-----|------------------------------------------|----------------|
| 0   | Documentation-only releases on main      | Manual tagging |
| 1   | Claude-generated code during generation  | GeneratorStop  |

When the user asks to tag a documentation-only release, use `v0.YYYYMMDD.N`. The orchestrator uses `v1.YYYYMMDD.N` when completing a generation (GeneratorStop): the merged code is tagged `v1.YYYYMMDD.N` and the requirements state is tagged `v1.YYYYMMDD.N-requirements`.

## Container Image Build

Every tag includes a container image build from `Dockerfile.claude`. When tagging, build and tag the image:

```bash
podman build -f Dockerfile.claude -t mage-claude-orchestrator:v[REL].YYYYMMDD.N .
```

The image tag matches the git tag. The `latest` tag is also applied to the most recent build.
