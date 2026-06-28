# Current Task

This file defines the active slice for the next local LLM session.

## Active Phase

D5C - scoring review, simulations, and balance tuning.

## Current Goal

Manually review the implemented D5C scoring and balance changes before marking D5C accepted. Any further balance change must be explicit, documented, and backed by tests, simulations, or written reasoning.

## Task Statement

D1, D2, D3, D4A, D4B, D4C, D4D, D4E, D5A, and D5B have been reviewed and accepted. D5B ANSI presentation is accepted after SyncTerm review, plain fallback sanity checks, final banking/army status fixes, standalone repository setup, and README/repository identity cleanup.

The next D5 slice is D5C: leaderboard scoring review, representative empire simulations, money score coefficient review, production/combat/action-cost balance review, and final balance decision documentation.

Do not add public deployment, firewall/systemd changes, live hub listing, live registration token use, destructive purge tooling, full visitor gameplay, or InterDoor protocol changes unless separately approved in a later phase.

## Required Inputs

- `PROJECT.md`
- `DECISIONS.md`
- `REWRITE_PLAN.md`
- `PHASE_PLAN.md`
- `SOURCE_NOTES.md`
- `TESTING.md`
- `CONFIG.DOM` for startup defaults
- `DOMINION.PAS` only if source-model fields or original region/building/mine concepts need confirmation
- `internal/game/*` for score, economy, military, production, turns, and action constants
- `internal/data/*` only if persistence affects score/report interpretation
- `internal/session/*` only if menu labels or report text must be adjusted for balance clarity

Outside-folder files may be read only for current InterDoor protocol/interface requirements, and only after stating why they are needed. D5C should not require outside-folder inspection.

## D5B Accepted Scope

- `-ansi` CLI flag controls ANSI color presentation.
- ANSI is enabled by default for CLI sessions and can be disabled with `-ansi=false`.
- `internal/session/presentation.go` now has an ANSI/plain presenter.
- ANSI mode uses native Empire Ascendant screens: compact header, framed panels, bright command keys, black background, and command strip styling.
- Menus use BBS-style boxed two-column panels.
- Plain fallback remains readable and contains no ANSI escapes.
- Banking, worker, develop, and attack blocked-action feedback is visible in ANSI mode.
- Army status and develop status use color-separated sections.
- `HEADER1.ANS` and the reusable menu frame are stored as Empire Ascendant runtime assets under `assets/ansi/`.
- The repository now has an Empire Ascendant README using `screenshot.png`, with old reference screenshots and unused sample ANSI files removed.
- D5B keeps command keys, mechanics, balance constants, database schema, federation behavior, deployment behavior, visitor gameplay, and purge tooling unchanged.

## D5C Implemented For Review

- Score weights are named constants.
- Leaderboard score now uses total wealth: `Money + MoneyBank`.
- Money scoring now uses the D5-planned `0.1` coefficient through `ScoreMoneyDivisor = 10`.
- Focused representative tests cover starting, builder/economic, and raider states.
- Balance tests cover early mining viability and default combat thresholds.
- Production constants, combat strengths, and action costs are retained pending user review.

## User Review Result

- ANSI visuals and aesthetics are accepted.
- Banking works after the final D5B fix pass.
- Army status color separation looks correct.
- Plain fallback sanity checks passed.

## Next D5C Acceptance Criteria

- Current score formula is documented and reviewed against representative empires.
- Representative economy and military states are simulated or tested.
- Money score coefficient is reviewed and either retained with rationale or changed with tests.
- Production, combat ratios, and action costs are reviewed together so changes do not create obvious dead ends.
- Any changed constants are centralized and covered by focused tests or documented simulation output.
- Final balance decisions are recorded in `DECISIONS.md` and `TESTING.md`.
- User manually confirms rankings/banking/economic-vs-military balance feels acceptable.
- No deployment, public hub, visitor gameplay, destructive purge, or InterDoor protocol behavior changes are introduced.

## D5B Final Verification

```bash
go test ./...
make build
make smoke
printf 'S\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-help.db -stdio -ansi=false
printf 'Q\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-ansi.db -stdio -ansi=true
./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:2324 -db var/empireascendant.db -ssh-host-key var/ssh_host_rsa -ansi=true -ssh-encoding=auto
```

## Resume Prompt

```text
Resume D5C manual review only. Work inside the Empire Ascendant folder. Read PROJECT.md, CURRENT_TASK.md, DECISIONS.md, TESTING.md, SESSION_HANDOFF.md, REWRITE_PLAN.md, PHASE_PLAN.md, SOURCE_NOTES.md, and the relevant internal/game files. D5B visuals are accepted. D5C scoring/balance implementation is pushed but not yet user-accepted. Review leaderboard scoring, banking score invariance, representative builder/raider balance, early mining viability, and default combat risk. Do not change deployment, live hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior. Ask before making additional balance changes unless a clear bug is found.
```

## Stop Conditions

Stop and ask before proceeding if:

- review or fixes require modifying files outside this folder
- implementation requires changing a shared InterDoor component
- implementation requires another new dependency
- balance changes are not backed by tests, simulations, or documented reasoning
- work starts changing deployment, public hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior
- the model starts following a sample game rather than Empire Ascendant's local docs
- deployment, firewall, public registration, or live hub token use is involved
