# D5B Visual Handoff

Date: 2026-06-25

## Status

D5B is paused. The generated ANSI implementation is technically working, but the visual design was rejected by user review and is not accepted.

## User Feedback

The user wants the game to visually feel much closer to the original Dominion screenshots: authored, classic BBS/door ANSI, not safe colored ASCII and not generated-looking panels.

The latest generated ANSI pass was rejected as looking amateur/childlike.

## Current Technical State

- `-ansi` exists and defaults to true.
- `-ansi=false` remains the safe plain fallback.
- `internal/session/presentation.go` contains the current generated ANSI presenter.
- Tests and build passed for the current technical implementation.
- Mechanics, balance, database schema, deployment, live hub behavior, visitor gameplay, and purge tooling were not changed by D5B.

## Preferred Next Direction

Use an authored TheDraw mockup as the visual source of truth for the ANSI main menu.

Recommended asset path:

```text
SCREENS/EA_MAIN.ANS
```

Acceptable alternatives inside this folder:

```text
assets/ansi/main_menu.ans
SCREENS/MAINMENU.ANS
```

The mockup should target classic 80x24 ANSI and include the command prompt area.

## Implementation Options To Plan After Mockup Exists

Option A: render the `.ANS` asset directly in `-ansi=true`, then append or overlay the live prompt using Go.

Option B: use the `.ANS` asset as the visual reference and recreate it in Go strings.

Prefer Option A if the exported `.ANS` displays correctly over Linux SSH/stdout without requiring a new parser dependency. Prefer Option B if direct rendering creates cursor-positioning or terminal compatibility issues.

## Guardrails

- Do not mark D5B accepted.
- Do not continue inventing generated ANSI screens without explicit user approval.
- Do not change gameplay mechanics, balance constants, database schema, deployment, live hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior.
- Work only inside the Empire Ascendant folder and subfolders.
- Do not follow Ledger of the Low.
- Outside-folder files are only allowed for current InterDoor protocol/interface facts, after stating why.

## Last Known Verification

These commands passed for the current technical implementation:

```bash
go test ./...
make build
make smoke
printf 'S\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-help.db -stdio -ansi=false
printf 'Q\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-ansi.db -stdio -ansi=true
timeout 2s ./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:0 -db var/empireascendant.db -ansi=true
```

The listener command exits with status `124` because `timeout` stops the long-running SSH listener after startup.

## Resume Prompt

```text
Resume D5B visual implementation only. Work inside the Empire Ascendant folder. Read PROJECT.md, CURRENT_TASK.md, DECISIONS.md, TESTING.md, SESSION_HANDOFF.md, and D5B_VISUAL_HANDOFF.md. The previous generated ANSI design was rejected and D5B is not accepted. Use the user-provided TheDraw mockup as the source of truth for the ANSI main menu. Keep `-ansi=false` as plain fallback. Do not change mechanics, balance constants, deployment, live hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior. Present a narrow file-level plan before coding unless the user explicitly approves implementation.
```
