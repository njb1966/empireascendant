# Empire Ascendant Decision Log

This file records settled decisions so local LLM sessions do not re-litigate them.

## Decision Format

Each decision should include:

- date
- decision
- reason
- rejected alternatives, if relevant
- affected files or phases

## Settled Decisions

### 2026-06-24 - Project Name

Decision: The game name is Empire Ascendant.

Reason: `REWRITE_PLAN.md` establishes this as the rewrite identity.

Rejected alternatives: Reusing Dominion as the public game title for the rewrite.

Affected phases: All phases.

### 2026-06-24 - Rewrite Type

Decision: This is a design-informed green-field rewrite, not a strict port.

Reason: `REWRITE_PLAN.md` notes that `Play_Game` in `DOMINION.PAS` is empty. The Pascal source provides data model, menu, config, and intent, but not complete gameplay.

Rejected alternatives: Treating the Pascal source as complete gameplay behavior.

Affected phases: All implementation phases.

### 2026-06-24 - Platform And Storage

Decision: Implement in Go using SQLite.

Reason: Empire Ascendant is planned as an InterDOOR module, and `REWRITE_PLAN.md` specifies SQLite-backed node state.

Rejected alternatives: ORM-first design, external database server, non-Go implementation.

Affected phases: D1-D5.

### 2026-06-24 - InterDOOR Module Shape

Decision: Build Empire Ascendant as a folder-local Go project that produces `interdoor-dominion`. The exact local package layout will be decided during D1 planning.

Reason: The user established this folder as the Empire Ascendant source of truth. Outside InterDoor material may constrain protocol/interface compatibility, but must not define this project's implementation shape.

Rejected alternatives: Following any sample game as a template, replacing another game node, or treating a parent/sibling project as the Empire Ascendant source of truth.

Affected phases: D1 and later.

### 2026-06-24 - Default Port

Decision: The default Empire Ascendant port is `2324`.

Reason: `REWRITE_PLAN.md` reserves `2324` for `interdoor-dominion`.

Rejected alternatives: Reusing `2323`, which is reserved for other local InterDoor work.

Affected phases: D1, deployment.

### 2026-06-24 - Startup Defaults Source

Decision: Use `CONFIG.DOM` as the source of truth for original startup defaults.

Reason: `CONFIG.DOM` is the parsed configuration file. The comment block in `DOMINION.PAS` is useful historical context but disagrees with the actual config file.

Rejected alternatives: Using the top comment in `DOMINION.PAS` as authoritative for defaults.

Affected phases: D1, D2, source archaeology.

### 2026-06-24 - Folder Source Of Truth

Decision: `/media/nick/1TB_Storage1/projects/retro/gaming/interdoor/games/empireascendant` is the source of truth for Empire Ascendant.

Reason: The user explicitly directed that work remain in this folder and subfolders. References outside this folder are copies or interface references only.

Rejected alternatives: Treating the parent InterDoor repository or any sibling game as authoritative for Empire Ascendant design or implementation.

Affected phases: All phases.


### 2026-06-24 - D1 Review Mode

Decision: D1 uses local terminal/stdio mode for live progress review. SSH listener support is deferred unless local review is insufficient.

Reason: The user only needs SSH early if there is no other practical way to review progress live. A local terminal loop reduces D1 complexity while preserving the path toward an SSH node later.

Rejected alternatives: Adding SSH listener complexity before the core game loop, persistence, and report flow exist.

Affected phases: D1, later node integration.

### 2026-06-24 - InterDoor Foundation Without D1 Friction

Decision: InterDoor is foundational to the project, but D1 should add only the local seams needed to avoid blocking later InterDoor integration. Register/heartbeat/federation behavior can be implemented later unless it proves trivial and non-disruptive.

Reason: Empire Ascendant exists to become an InterDoor node, but early momentum depends on proving the game skeleton first.

Rejected alternatives: Ignoring InterDoor until the end, or forcing federation into D1 before the local game is reviewable.

Affected phases: D1-D4.


### 2026-06-24 - D1 Accepted

Decision: D1 local terminal walking skeleton is accepted and the project may proceed to D2 planning.

Reason: User review found no issues. Verification passed for `go test ./...`, `make build`, and local stdio smoke flow.

Rejected alternatives: Extending D1 with SSH listener or federation before D2.

Affected phases: D1, D2.

### 2026-06-24 - D2 Economy Constants

Decision: D2 uses named balance constants for economy values not fully specified by the original source: default fossil plants, mineral left per mine, mine prices, build costs, research costs, worker hire cost, and mine purchase cost.

Reason: `CONFIG.DOM` and `DOMINION.PAS` provide starting concepts and some defaults, but not all D2 balance numbers. Central constants keep the values explicit and easy to tune.

Rejected alternatives: Hiding balance values in menu code or SQL, or claiming unspecified constants came from original Dominion.

Affected phases: D2, D5 balance.

### 2026-06-24 - D2 Implemented For Review

Decision: D2 economy implementation is complete enough for user review.

Reason: Automated tests and manual smoke verification pass. D2 remains scoped to resource production and non-combat empire development.

Rejected alternatives: Adding SSH, federation, combat, travel, or polish before D2 review.

Affected phases: D2, D3.

### 2026-06-24 - D2 Backfills Existing D1 Empires

Decision: Loading economy state backfills D2 economy rows for existing D1-created empires that do not yet have `empire_regions`, `empire_buildings`, `empire_mines`, or `empire_tech` rows.

Reason: The local review database already contained D1 empire records. D2 must migrate those records in place rather than failing with missing economy rows.

Rejected alternatives: Requiring the user to delete the local database, manually recreate empires, or run a separate migration command for this early local phase.

Affected phases: D2.

### 2026-06-24 - D2 Accepted

Decision: D2 economy implementation is accepted and the project may proceed to D3 planning.

Reason: User review confirmed the implementation is working. Verification passed for `go test ./...`, `make build`, `make smoke`, and local database backfill.

Rejected alternatives: Extending D2 with aesthetics, SSH listener, federation, or combat before D3 planning.

Affected phases: D2, D3.

### 2026-06-24 - D3 Local Military Scope

Decision: D3 implements local same-node military behavior only: military persistence, recruitment, defenses, military research, ground assault, ballistic strike, attack limits, dispatches, and spy missions.

Reason: `PHASE_PLAN.md` scopes D3 to local combat and explicitly defers federation, cross-node travel, news, and polish. The user approved D3 after confirming SSH and InterDoor integration can wait if they do not block game progress.

Rejected alternatives: Adding SSH listener, InterDoor federation, cross-node attacks, Galactic News, or UI polish in D3.

Affected phases: D3, D4, D5.

### 2026-06-24 - D3 Balance Constants

Decision: D3 combat strengths, military costs, research costs, and attack limits are explicit constants in the game layer.

Reason: The original Pascal source establishes broad military concepts but does not provide a complete combat loop. Explicit constants keep D3 playable and testable while leaving final tuning to D5.

Rejected alternatives: Hiding balance values in SQL or terminal menu code, or claiming D3 balance values came from a complete original combat implementation.

Affected phases: D3, D5.

### 2026-06-24 - D3 Implemented For Review

Decision: D3 military implementation is complete enough for user review.

Reason: Automated tests, build, smoke target, and a manual stdio run pass. D3 remains local-only and does not add federation, cross-node travel, Galactic News, or polish.

Rejected alternatives: Proceeding into D4 before D3 review, or extending D3 with network behavior.

Affected phases: D3, D4.

### 2026-06-24 - D3 Accepted

Decision: D3 military implementation is accepted and the project may proceed to D4A InterDoor foundation implementation.

Reason: User review confirmed D3 looks good.

Rejected alternatives: Continuing to tune D3 balance or presentation before adding InterDoor foundation.

Affected phases: D3, D4.

### 2026-06-24 - D4A Foundation Before Cross-Node Mechanics

Decision: D4 starts with InterDoor foundation only: registration, heartbeat, local event log, event push/pull, roster push/pull, Rankings, Galactic News, Wanderers, and fake-hub tests.

Reason: Registration, auth, event cursors, and roster sync are prerequisites for reliable cross-node combat and travel. Keeping this slice separate reduces risk and preserves local play during hub outages.

Rejected alternatives: Implementing cross-node PvP, travel, SSH listener, or deployment in the same slice.

Affected phases: D4A, D4B.

### 2026-06-24 - Empire Ascendant InterDoor Identity

Decision: Empire Ascendant uses `empire_ascendant` as `game_id`, `Empire Ascendant` as `game_title`, protocol version `1`, and `empireascendant.*` event names for game-defined events.

Reason: `game_id` is a stable protocol identifier and should not reuse another game's identity. Event names should describe Empire Ascendant, not the original Dominion working title or any sample game.

Rejected alternatives: `ledger_of_the_low`, `dominion.*`, or using human display text as the protocol identifier.

Affected phases: D4A and later federation features.

### 2026-06-24 - D4A Implemented For Review

Decision: D4A InterDoor foundation is complete enough for review.

Reason: Automated tests, fake-hub tests, build, smoke target, and a manual stdio run pass. D4A remains foundation-only and does not add cross-node PvP, travel, SSH listener, or deployment.

Rejected alternatives: Using the live hub token during automated tests, or proceeding into D4B before D4A review.

Affected phases: D4A, D4B.

### 2026-06-25 - D4B Cross-Node PvP Scope

Decision: D4B implements cross-node PvP only: remote target selection, hub PvP queue calls, victim-node pending drain, victim-side ground/missile resolution, `pvp.resolved` events, attacker-side result application, and malformed-payload blocking.

Reason: D4A established registration, event, and roster foundation. Cross-node PvP is the next InterDoor mechanic that can be built without travel, SSH listener, deployment, or presentation polish.

Rejected alternatives: Adding travel, SSH listener, deployment, live hub token use, or D5 visual polish during the D4B slice.

Affected phases: D4B, D4C, D5.

### 2026-06-25 - D4B Victim-Side Resolution

Decision: Remote combat is queued by the attacker node but resolved only by the victim node. The attacker node applies only the attacker-side effects described by the resulting `pvp.resolved` event.

Reason: This matches the current InterDoor PvP queue contract and prevents one node from mutating another node's local empire state.

Rejected alternatives: Resolving remote victim combat on the attacker node, or requiring the victim node to mutate attacker state directly.

Affected phases: D4B and later federation mechanics.

### 2026-06-25 - D4B Implemented For Review

Decision: D4B cross-node PvP is complete enough for user review.

Reason: The implementation adds focused client, Store, Syncer, and session coverage while keeping travel, SSH listener, deployment, and polish out of scope. Automated tests pass with fake hubs only.

Rejected alternatives: Using the live hub token during automated tests, or combining D4B with travel/deployment.

Affected phases: D4B, D4C.

### 2026-06-25 - D4C Hyperdrive Travel Foundation Scope

Decision: D4C implements travel foundation only: local empire snapshot export, hub travel submission, local traveling state, pending arrival drain, destination snapshot import, `player.traveled` events, malformed-arrival blocking, and home-node login restriction while away.

Reason: D4B completed cross-node PvP. Travel is the remaining InterDoor federation mechanic in D4 that can be implemented locally with fake-hub coverage before SSH listener or deployment.

Rejected alternatives: Adding SSH listener, deployment, live hub registration, full visitor gameplay, inactive empire lifecycle, or D5 polish during the D4C slice.

Affected phases: D4C, D4D, D5.

### 2026-06-25 - D4C Visitor Handling

Decision: D4C stores arriving away-from-home empires as visitor snapshots instead of enabling full visitor gameplay.

Reason: SSH listener and visitor-session routing are still deferred. Storing the imported snapshot verifies the travel queue and arrival mechanics without creating a second local gameplay mode prematurely.

Rejected alternatives: Importing visitors directly into local `empires` as playable accounts, or blocking all non-home arrivals until SSH listener work exists.

Affected phases: D4C and later visitor gameplay.

### 2026-06-25 - D4C Implemented For Review

Decision: D4C Hyperdrive travel foundation is complete enough for user review.

Reason: Automated tests cover travel wire shape, local travel state, snapshot import/idempotency, malformed blocking, sync fake-hub travel drain, and terminal Hyperdrive submission. Live hub token use and deployment remain out of scope.

Rejected alternatives: Using the live hub token during automated tests, or combining D4C with SSH/deployment.

Affected phases: D4C, D4D.

### 2026-06-25 - Case-Insensitive Usernames

Decision: Usernames are case-insensitive for lookup and duplicate prevention, while preserving the player's chosen casing for display/storage.

Reason: Terminal users may vary casing between sessions. Treating `CalvusRex`, `calvusrex`, and `CALVUSREX` as different users is confusing and can accidentally create duplicate empires.

Rejected alternatives: Forcing all usernames to lowercase visually, or leaving login case-sensitive.

Affected phases: D1 and all later login/session flows.

### 2026-06-25 - D4D Inactive Empire Lifecycle

Decision: Inactive empires move through `active`, `warned`, `hidden`, and `purge_eligible` states based on last login. Hidden and purge-eligible empires are removed from public rankings, roster push, heartbeat count, and local attack target selection, but are not deleted.

Reason: Public federation lists and local target lists should not fill with abandoned review/test empires. Deletion is destructive and should remain an explicit operator decision outside this slice.

Rejected alternatives: Physically deleting old empires automatically, or leaving inactive empires visible indefinitely.

Affected phases: D4D and later operations tooling.

### 2026-06-25 - D4D Implemented For Review

Decision: D4D inactive empire lifecycle is complete enough for user review.

Reason: Automated tests cover lifecycle transitions, non-destructive purge eligibility, visibility filtering, and login reactivation. Build, smoke, and manual stdio smoke pass.

Rejected alternatives: Combining lifecycle with SSH listener, deployment, or destructive purge tooling.

Affected phases: D4D, D4E.

### 2026-06-25 - D4E Local SSH Listener Foundation

Decision: D4E adds a local SSH listener that adapts SSH session channels into the existing Empire Ascendant terminal runner.

Reason: Live local review through SSH is useful before public deployment, but the game loop should remain single-sourced through `session.Runner`. Stdio remains the default review mode, and SSH listener mode is explicitly enabled with `-stdio=false`.

Rejected alternatives: Implementing public deployment, firewall/systemd changes, live hub listing, or a separate SSH-only gameplay path during this slice.

Affected phases: D4E, deployment.

### 2026-06-25 - D4E Transport Authentication Boundary

Decision: D4E SSH password authentication only admits a client into the local SSH shell session. Empire Ascendant's in-game username/password remains the authoritative player identity.

Reason: Binding SSH users to game users is a later operations/design question. Keeping the boundary simple allows local review without changing existing account behavior.

Rejected alternatives: Mapping SSH usernames to empire usernames, adding public account provisioning, or bypassing the in-game login prompt.

Affected phases: D4E and later deployment/account policy.

### 2026-06-25 - D4E Accepted

Decision: D4E local SSH listener foundation is accepted and the project may proceed to D5A planning.

Reason: Automated tests, build, smoke, listener startup, and manual SSH review passed after fixing SSH terminal input echo and line handling. The user confirmed SSH testing appears to work.

Rejected alternatives: Starting deployment, live hub listing, full visitor gameplay, or balance tuning before planning D5 polish.

Affected phases: D4E, D5A.

### 2026-06-25 - D5A Terminal Presentation Scope

Decision: D5A uses small ASCII terminal presentation helpers and Story / Instructions content instead of importing or rendering the original ANSI files directly.

Reason: The local ANSI assets and screenshots establish the intended BBS feel, but the current review paths must remain readable through stdio and SSH without ANSI parsing. ASCII section/menu helpers improve consistency while preserving terminal simplicity and keeping mechanics unchanged.

Rejected alternatives: Adding ANSI parsing/rendering, changing game mechanics during polish, adding dependencies, or copying obsolete original menu structure wholesale.

Affected phases: D5A, later D5 polish.

### 2026-06-25 - D5A Implemented For Review

Decision: D5A terminal presentation, Story / Instructions, and basic menu readability are complete enough for user review.

Reason: The implementation is limited to presentation helpers, menu rendering calls, and instructions text. It keeps balance, persistence, federation, deployment, visitor gameplay, and purge tooling out of scope.

Rejected alternatives: Combining D5A with balance tuning or deployment.

Affected phases: D5A, D5B.

### 2026-06-25 - D5A Accepted

Decision: D5A terminal presentation, Story / Instructions, and basic menu readability are accepted.

Reason: Terminal output was reviewed. The user accepted D5A and approved moving forward.

Rejected alternatives: Continuing to adjust D5A before planning the stronger aesthetics slice.

Affected phases: D5A, D5B.

### 2026-06-25 - D5B Native BBS Presentation

Decision: D5B implements native Empire Ascendant BBS-style presentation with optional ANSI color and plain fallback, rather than rendering the original `.ANS` files directly.

Reason: Local ANSI files and screenshots are useful visual references, but their menu labels and cursor-positioning are tied to unfinished original Dominion screens. A native presenter keeps current Empire Ascendant command flow intact, works over stdio and SSH, and avoids adding an ANSI parsing dependency.

Rejected alternatives: Raw `.ANS` rendering, new ANSI parsing dependency, changing command keys, or changing mechanics during visual polish.

Affected phases: D5B, later D5 polish.

### 2026-06-25 - D5B Generated ANSI Rejected

Decision: D5B is paused pending an authored visual mockup. The generated ANSI presenter is technically working but not accepted as the final visual direction.

Reason: User review found the generated ANSI screen unacceptable and amateur-looking. ANSI/BBS aesthetics should be treated as authored screen art, not guessed layout. The next pass should use a user-created TheDraw mockup or another explicitly approved design source.

Rejected alternatives: Continuing to iterate on generated ANSI art without a mockup, marking D5B accepted because tests pass, or moving to the next phase before visual direction is resolved.

Affected phases: D5B.

### 2026-06-25 - D5B Implemented For Review

Decision: D5B ANSI/title art and BBS-style screen presentation are complete enough for user review.

Reason: Automated tests, build, smoke, plain stdio, ANSI stdio, and ANSI-enabled SSH listener startup pass. D5B remains presentation-only.

Rejected alternatives: Combining D5B with balance tuning, deployment, or raw ANSI file rendering.

Affected phases: D5B, D5C.

### 2026-06-25 - D5B ANSI Presentation Revised

Decision: D5B ANSI mode is the primary player-facing presentation and targets classic 80x24 BBS door aesthetics, with plain ASCII retained only as a fallback.

Reason: User review found the first native presenter too safe and dull compared with the local Dominion screenshots. The revised presenter keeps current Empire Ascendant command flow but uses denser screenshot-informed framing, black-background ANSI panels, bright command keys, and compact screen layout.

Rejected alternatives: Treating ANSI as a minor color overlay, keeping the oversized splash-style title, rendering raw `.ANS` files before the current command flow is stable, or changing mechanics during visual revision.

Affected phases: D5B, later D5 polish.
