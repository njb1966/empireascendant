# Empire Ascendant Project Control

This file is the regular grounding document for local LLM work on Empire Ascendant. Load it at the start of every session.

## Project Identity

Empire Ascendant is a Go rewrite of the unfinished Turbo Pascal BBS door game Dominion. This folder is the source of truth for the Empire Ascendant project:

`/media/nick/1TB_Storage1/projects/retro/gaming/interdoor/games/empireascendant`

Empire Ascendant is intended to speak InterDoor where needed, but it is not a clone of any sample InterDoor game. Other InterDoor projects may be read only to confirm current protocol or interface requirements. They are not design authorities for Empire Ascendant.

This is a design-informed rewrite, not a line-by-line port. The original Pascal source contains the data model, configuration format, player creation, menus, and authorial intent, but the core gameplay procedure `Play_Game` is empty.

## Folder Boundary

Work inside this folder and its subfolders only.

Files outside this folder are reference material only. Do not modify outside files unless the user explicitly approves a separate task. Do not plan Empire Ascendant around another game's structure, depth, tone, or implementation choices.

If current InterDoor integration requirements must be checked outside this folder, state exactly which requirement is being checked and treat the result as an interface constraint only.

## Hard Constraints

- Language: Go.
- Database: SQLite.
- Game name: Empire Ascendant.
- This folder is the project root and source of truth.
- Build the project from this folder and subfolders unless a specific InterDoor interface check requires looking outside.
- Binary name: `interdoor-dominion` unless superseded by a later local decision.
- Default local game port: `2324`.
- Preserve terminal/BBS simplicity.
- Keep changes small, explicit, and phase-scoped.
- Prefer Empire Ascendant's local architecture and project needs over sample-game patterns.
- Do not add dependencies unless the task clearly requires them.

## Source Precedence

When sources disagree, use this order:

1. `PROJECT.md` for current project control.
2. `DECISIONS.md` for settled decisions.
3. `CURRENT_TASK.md` for the active slice.
4. `REWRITE_PLAN.md` for intended Go rewrite behavior.
5. `PHASE_PLAN.md` for phase scope.
6. `CONFIG.DOM` for original startup defaults.
7. `DOMINION.PAS` and `SETUP.PAS` for original data model, config parsing, player creation, menus, and historical intent.
8. `SOURCE_NOTES.md` for source archaeology summaries.
9. `INTERDOOR_DESIGN_GUIDE.md` for broad InterDoor context only.
10. `README.md` for current standalone game usage.

Outside-folder files never outrank the local project files above.

If a task depends on a conflict between sources, stop and identify the conflict before implementing.

## Current Phase

D5C - scoring review, simulations, and balance tuning plan.

## Done

- Original Pascal source is present: `DOMINION.PAS`, `SETUP.PAS`, `A.PAS`.
- Original config is present: `CONFIG.DOM`.
- Empire Ascendant ANSI runtime assets are present in `assets/ansi/`.
- Current game usage is documented in `README.md`.
- InterDoor context is in `INTERDOOR_DESIGN_GUIDE.md`.
- Rewrite intent is in `REWRITE_PLAN.md`.
- Project control and LLM workflow docs are present.
- Folder-local source-of-truth boundary is established.
- D1 local Go skeleton is implemented with stdio review mode, SQLite persistence, player creation, duplicate checks, daily turn reset, HQ shell, and Empire Report.
- D1 was reviewed and accepted with no issues found.
- D2 economy is implemented with seeded regions/buildings/mines/tech, daily production, bank/withdraw, region activation, building/research/mine/worker/mineral actions, Develop menu, expanded report, and focused tests.
- D2 was reviewed and accepted.
- D3 military is implemented with local military persistence, recruitment, defense building, military research, local ground assault, same-node ballistic strike, attack limits, dispatch inbox, spy scout/sabotage missions, expanded report, and focused tests.
- D3 was reviewed and accepted.
- D4A InterDoor foundation is implemented with federation config flags, HTTP client, registration/heartbeat/sync one-shot commands, local event log, event push/pull, roster push/pull, Rankings, Galactic News, Wanderers display, and fake-hub tests.
- D4A was reviewed and accepted.
- D4B cross-node PvP is implemented with remote target selection, hub PvP queue calls, local outbound tracking, victim-node pending drain, victim-side ground/missile resolution, `pvp.resolved` events, attacker-side result application, malformed-payload blocking, and focused fake-hub tests.
- D4B was reviewed and accepted.
- D4C Hyperdrive travel foundation is implemented with travel wire calls, local travel state, snapshot export/import, home-node login restriction while away, pending arrival drain, visitor snapshot storage, `player.traveled` events, malformed-arrival blocking, and focused fake-hub tests.
- D4C was reviewed and accepted.
- D4D inactive empire lifecycle is implemented with last-login tracking, warned/hidden/purge-eligible statuses, login reactivation, non-destructive purge eligibility, and filtering from rankings, roster, heartbeat count, and local attack targets.
- D4D was reviewed and accepted.
- D4E local SSH listener foundation is implemented with `-stdio=false`, configurable listen address, persistent Ed25519 host key, password-gated SSH shell sessions, minimal SSH terminal echo/line editing, and session routing into the existing Empire Ascendant terminal runner.
- D4E was reviewed and accepted after SSH connection, terminal input, and game session access were verified manually.
- D5A terminal presentation, Story / Instructions, and basic menu readability are implemented with ASCII terminal helpers, no balance changes, no mechanics changes, and no deployment behavior.
- D5A was reviewed and accepted.
- D5B ANSI presentation is accepted after user review in SyncTerm and plain fallback review. ANSI mode is the primary player-facing UI, plain mode remains a readable fallback, and the standalone repository identity cleanup was completed.

## Next Task

Plan and execute D5C: leaderboard scoring review, representative empire simulations, money score coefficient review, production/combat/action-cost balance review, and final balance decision documentation. Deployment, live hub listing, full visitor gameplay, and destructive purge tooling remain deferred until separately planned and approved.

## Do Not

- Do NOT rename Empire Ascendant.
- Do NOT treat this as a new unrelated game.
- Do NOT claim Dominion had a complete gameplay loop; `Play_Game` is empty.
- Do NOT use `DOMINION.PAS` comments over `CONFIG.DOM` when resolving startup defaults.
- Do NOT follow any sample game as an implementation template.
- Do NOT make another game the standard for Empire Ascendant's depth, architecture, tone, or scope.
- Do NOT modify files outside this folder without explicit approval.
- Do NOT introduce an ORM unless explicitly approved.
- Do NOT replace SQLite.
- Do NOT introduce Docker, Kubernetes, cloud deployment layers, a web UI, or SPA architecture unless explicitly requested.
- Do NOT implement features from a later phase while working on an earlier phase.
- Do NOT make broad refactors merely to modernize the project.
- Do NOT assume deployment changes are safe; infrastructure changes require explicit review and must respect the infrastructure source of truth.

## Required LLM Behavior

Before writing code, the LLM must restate:

- current phase
- exact task
- files it expects to inspect
- files it expects to edit
- whether any outside-folder reference is required and why
- constraints that apply
- source evidence it is relying on
- acceptance criteria

If the model cannot identify source evidence, it should ask for clarification or inspect the relevant local files instead of guessing.

## Session Rule

Use one implementation slice per session. If the conversation gets long, the model starts inventing behavior, follows a sample game, or stops citing project documents correctly, clear the session and re-ground from the files.
