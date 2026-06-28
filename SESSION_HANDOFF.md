# Session Handoff

Date: 2026-06-28

## Current State

- Standalone repository: `njb1966/empireascendant`.
- Current branch: `main`.
- Latest pushed commit at handoff: `e61d15e` (`Implement D5C scoring balance review`).
- D1, D2, D3, D4A, D4B, D4C, D4D, D4E, D5A, and D5B are implemented, reviewed, and accepted.
- D5B ANSI presentation is accepted after SyncTerm review, plain fallback sanity checks, banking/feedback fixes, and repository identity cleanup.
- D5C scoring/balance implementation is complete and pushed, but still needs user manual review before marking D5C accepted.

## D5C Implemented Scope

- Score weights are named constants in `internal/game/score.go`.
- Leaderboard score now uses total wealth: `Money + MoneyBank`.
- Money scoring now uses the D5-planned `0.1` coefficient through `ScoreMoneyDivisor = 10`.
- Focused scoring tests cover:
  - formula components
  - free banking score invariance
  - representative starting, builder/economic, and raider states
- Focused balance tests cover:
  - early mining path viability
  - default combat threshold behavior
- Production constants, combat strengths, and action costs were retained.

## Verified Commands

Last known good verification:

```bash
go test ./internal/game
go test ./...
make smoke
```

## Manual Review Needed Next

Before marking D5C accepted, review enough gameplay to confirm the leaderboard and balance feel sane:

1. Create or reuse a few empires.
2. Compare rankings before and after banking money.
3. Confirm depositing money no longer lowers rank.
4. Confirm wealthy/economic empires feel competitive with military-heavy empires.
5. Confirm early development still feels reasonable:
   - activate region
   - build toward Miners Guild
   - hire/assign miner
   - sell minerals
6. Confirm combat still feels risky and not obviously profitable against a fresh/default empire.

## Next Prompt

```text
Resume Empire Ascendant D5C manual review. Work only inside /media/nick/1TB_Storage1/projects/retro/gaming/interdoor/games/empireascendant. Read PROJECT.md, CURRENT_TASK.md, DECISIONS.md, TESTING.md, SESSION_HANDOFF.md, REWRITE_PLAN.md, PHASE_PLAN.md, and SOURCE_NOTES.md. D5B is accepted. D5C scoring/balance implementation is pushed at e61d15e but not yet user-accepted. Help review leaderboard scoring, banking score invariance, representative builder/raider balance, early mining viability, and default combat risk. Do not change deployment, live hub behavior, visitor gameplay, purge tooling, or InterDoor protocol behavior unless separately approved. Ask before making additional balance changes unless a clear bug is found.
```

## Guardrails

- Work only inside this folder and subfolders unless the user explicitly approves a separate task.
- Do not inspect outside-folder files unless confirming a current InterDoor protocol/interface fact, and state why first.
- Do not mark D5C accepted until the user reviews the scoring/balance behavior.
- Do not start deployment, firewall/systemd changes, public registration, live hub token use, full visitor gameplay, or destructive purge tooling during D5C review.
- Any further balance change must be backed by tests, simulation output, or documented reasoning.

## Likely Paths After Review

- If manual D5C review passes: mark D5C accepted in `PROJECT.md`, `CURRENT_TASK.md`, `DECISIONS.md`, and `TESTING.md`.
- Then choose between deployment planning or the next gameplay slice.
