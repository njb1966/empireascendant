# Session Handoff

Date: 2026-06-25

## Current State

- D1, D2, D3, D4A, D4B, D4C, D4D, and D4E are implemented, reviewed, and accepted.
- D5A terminal presentation, Story / Instructions, and menu readability are implemented, reviewed, and accepted.
- D5B generated ANSI presentation is technically implemented and verified, but user review rejected the look. D5B is not accepted.
- User may create an authored TheDraw mockup to use as the D5B visual source of truth.
- Legacy `var/empireascendant.db` username migration was fixed and verified after `username_key` index creation failed on an older schema.
- D4E SSH listener was manually verified: SSH connects, terminal input works, and the game session is reachable.
- The SSH review database caveat is documented: `/tmp/empireascendant-d4e-ssh.db` is throwaway and will not contain empires from `var/empireascendant.db`.
- Active next task is waiting for the authored TheDraw mockup or an explicit user request for another design attempt.

## Next Prompt

```text
Resume D5B visual implementation only. Work inside the Empire Ascendant folder. Read PROJECT.md, CURRENT_TASK.md, DECISIONS.md, TESTING.md, SESSION_HANDOFF.md, and D5B_VISUAL_HANDOFF.md. The previous generated ANSI design was rejected and D5B is not accepted. Use the user-provided TheDraw mockup as the source of truth for the ANSI main menu. Keep `-ansi=false` as plain fallback. Do not change mechanics, balance constants, deployment, live hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior. Present a narrow file-level plan before coding unless the user explicitly approves implementation.
```

## D5A Accepted Scope

- ASCII terminal title, section, and menu helpers.
- Story / Instructions from the main menu.
- Consistent formatting for main menu, Empire HQ, Develop, Attack, Rankings, News, Wanderers, and Dispatches.
- No mechanics, balance, deployment, federation, visitor gameplay, or purge tooling changes.

## D5B Implemented Scope

- `-ansi` CLI flag.
- ANSI primary mode and plain fallback.
- Compact native Empire Ascendant header.
- Black-background framed menu panels with bright command keys.
- Presentation-only work; no mechanics or balance changes.

## D5B User Review Result

- Generated ANSI design was rejected as visually unacceptable.
- Do not continue generated ANSI art iteration by default.
- Next preferred direction is an authored TheDraw `.ANS` mockup used as the primary ANSI visual target.
- Raw folder-local `.ANS` rendering is allowed to be considered for D5B if it best preserves the authored screen.

## Hard Guardrails

- Work only inside this folder and subfolders.
- Do not use Ledger of the Low as a design or implementation guide.
- Do not inspect outside-folder files unless confirming a current InterDoor protocol/interface fact, and state why first.
- Do not mark D5B accepted until the authored/mockup-based terminal output is reviewed and approved.
- Do not change balance, mechanics, deployment, public hub behavior, full visitor gameplay, or purge tooling during D5B review.

## Verification Already Good

Last known good verification:

```bash
go test ./...
make build
make smoke
printf 'S\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-help.db -stdio -ansi=false
timeout 2s ./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:0 -db var/empireascendant.db
printf 'Q\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-ansi.db -stdio -ansi=true
```

Manual SSH review command:

```bash
./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:2324 -db var/empireascendant.db -ansi=true
ssh -p 2324 review@127.0.0.1
```
