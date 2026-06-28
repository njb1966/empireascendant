# Current Task

This file defines the active slice for the next local LLM session.

## Active Phase

D5B - paused pending authored TheDraw ANSI mockup.

## Current Goal

Resume D5B only after the user provides an authored TheDraw ANSI mockup or explicitly asks for another design attempt.

## Task Statement

D1, D2, D3, D4A, D4B, D4C, D4D, D4E, and D5A have been reviewed and accepted. D5B generated ANSI presentation was implemented and verified technically, but user review rejected the look as amateur/childlike. D5B is not accepted.

The next D5B step is to use an authored TheDraw mockup as the visual source. Do not keep inventing generated ANSI screen designs unless the user explicitly asks for that.

Do not add public deployment, firewall/systemd changes, live hub listing, live registration token use, destructive purge tooling, full visitor gameplay, mechanics changes, or balance changes unless separately approved in a later phase.

## Required Inputs

- `PROJECT.md`
- `DECISIONS.md`
- `REWRITE_PLAN.md`
- `PHASE_PLAN.md`
- `SOURCE_NOTES.md`
- `TESTING.md`
- `CONFIG.DOM` for startup defaults
- `DOMINION.PAS` only if source-model fields or original region/building/mine concepts need confirmation
- `assets/ansi/*.ans` for local runtime ANSI assets

Outside-folder files may be read only for current InterDoor protocol/interface requirements, and only after stating why they are needed. D5B review should not require outside-folder inspection.

## D5B Implemented Scope

- `-ansi` CLI flag controls ANSI color presentation.
- ANSI is enabled by default for CLI sessions and can be disabled with `-ansi=false`.
- `internal/session/presentation.go` now has an ANSI/plain presenter.
- ANSI mode uses native Empire Ascendant screens: compact header, framed panels, bright command keys, black background, and command strip styling.
- Menus use BBS-style boxed two-column panels.
- Plain fallback remains readable and contains no ANSI escapes.
- D5B keeps command keys, mechanics, balance constants, database schema, federation behavior, deployment behavior, visitor gameplay, and purge tooling unchanged.

## User Review Result

- The generated ANSI design is not acceptable.
- User described it as looking like a child designed it.
- User may create a TheDraw mockup for the project to use as the visual target.
- Preferred direction is authored ANSI art for the primary mode and generated ASCII/plain fallback for `-ansi=false`.
- Do not mark D5B accepted until the authored visual direction is implemented and reviewed.

## Next D5B Acceptance Criteria

- Existing stdio and SSH play still work.
- Existing command keys and menu behavior are preserved.
- `-ansi=false` produces readable output without ANSI escape sequences.
- `-ansi=true` uses the user-approved authored ANSI visual direction.
- Main menu matches the approved TheDraw mockup closely enough for user review.
- No mechanics or balance constants changed.
- No dependency was added.
- D5B may render folder-local `.ANS` assets if that is the best implementation path.
- D5B must not rely on outside-folder design examples.

## Verified Commands

```bash
go test ./...
make build
make smoke
printf 'S\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-help.db -stdio -ansi=false
printf 'Q\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-ansi.db -stdio -ansi=true
timeout 2s ./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:0 -db var/empireascendant.db -ansi=true
```

Optional manual SSH review:

```bash
./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:2324 -db var/empireascendant.db -ansi=true
ssh -p 2324 review@127.0.0.1
```

D4E user review confirmed SSH connects, terminal input works, and the game session is reachable. Use `-ansi=false` when reviewing the plain fallback.

## Resume Prompt

```text
Resume D5B visual implementation only. Work inside the Empire Ascendant folder. Read PROJECT.md, CURRENT_TASK.md, DECISIONS.md, TESTING.md, SESSION_HANDOFF.md, and D5B_VISUAL_HANDOFF.md. The previous generated ANSI design was rejected and D5B is not accepted. Use the user-provided TheDraw mockup as the source of truth for the ANSI main menu. Keep `-ansi=false` as plain fallback. Do not change mechanics, balance constants, deployment, live hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior. Present a narrow file-level plan before coding unless the user explicitly approves implementation.
```

## Stop Conditions

Stop and ask before proceeding if:

- review or fixes require modifying files outside this folder
- implementation requires changing a shared InterDoor component
- implementation requires another new dependency
- review or fixes start changing mechanics, balance constants, deployment, or public hub behavior
- the model starts following a sample game rather than Empire Ascendant's local docs
- deployment, firewall, public registration, or live hub token use is involved
