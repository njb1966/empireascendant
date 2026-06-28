# Empire Ascendant Phase Plan

This phase plan expands the phase breakdown from `REWRITE_PLAN.md` into work slices suitable for local LLM sessions.

## Phase Rules

- Complete one slice per session when possible.
- Each slice must have acceptance criteria and verification.
- Do not implement later-phase features early.
- Changes outside this folder are out of scope unless explicitly approved. Outside InterDoor material may be inspected only for current protocol/interface requirements.
- Deployment work is separate from feature implementation.

## D0 - Documentation Control

Goal: Make project truth explicit so local LLM sessions stay grounded.

Slices:

1. Create project control docs.
2. Resolve source precedence.
3. Record known source conflicts.
4. Define session prompts and completion workflow.

Acceptance:

- `PROJECT.md` exists.
- `DECISIONS.md` exists.
- `CURRENT_TASK.md` exists.
- `SOURCE_NOTES.md` exists.
- `TESTING.md` exists.
- `STANDUP_TO_COMPLETION.md` exists.
- `PHASE_PLAN.md` exists.
- `LLM_SESSION_STEPS.md` exists.
- Existing docs point to the control workflow.

## D1 - Walking Skeleton

Goal: Empire Ascendant can start, create a player, persist D1 state, and display a basic report.

Slices:

1. Inspect local Empire Ascendant docs and source files.
2. Confirm any current InterDoor protocol/interface requirements that affect D1.
3. Decide the local Go project/package layout inside this folder.
4. Create the local `interdoor-dominion` command entrypoint.
5. Add D1 SQLite schema for `empires`.
6. Implement player creation with World Name and Empire Name.
7. Enforce duplicate World Name and Empire Name checks.
8. Implement daily turn reset using Unix day number.
9. Implement main menu shell and Empire HQ shell.
10. Implement `[T] Empire Report` for D1 resources.
11. Add focused tests and build verification.

Acceptance:

- `interdoor-dominion` builds from this folder.
- D1 schema initializes cleanly.
- Player creation works and persists.
- Duplicate names are rejected.
- Turn reset is tested.
- Empire Report shows D1 resource state.
- Local terminal/session path reaches the game shell end to end.
- Economy, military, federation, travel, and polish are still stubs or absent.

## D2 - Economy

Goal: Add resource production and non-combat empire development.

Slices:

1. Add `empire_regions`, `empire_buildings`, `empire_mines`, and `empire_tech` tables.
2. Seed starting region/building/mine/tech state from phase-approved defaults.
3. Implement daily food production and consumption.
4. Implement population growth and shrinkage.
5. Implement energy production.
6. Implement mine yield and depletion.
7. Implement research point and building point production.
8. Implement bank and withdraw actions.
9. Implement Develop Empire menu for activating regions.
10. Implement building labs/factories.
11. Implement energy tech research.
12. Implement mineral selling.
13. Add tests for daily production and resource changes.

Acceptance:

- Daily tick is deterministic under tests.
- Resource changes are persisted.
- Actions that cost turns decrement turns exactly once.
- Free actions do not decrement turns.
- D2 menus remain terminal-simple.

## D3 - Military

Goal: Add local combat, defenses, ballistic attacks, spy missions, and offline dispatches.

Slices:

1. Add `empire_military` table.
2. Add local attack/dispatch tables if not already present.
3. Implement recruitment menu.
4. Implement defense construction.
5. Implement weapons tech research.
6. Implement same-node ground assault.
7. Implement same-node ballistic strike.
8. Implement target/day attack limits.
9. Implement Galactic Dispatches inbox.
10. Implement spy scout and sabotage missions.
11. Add combat and limit tests.

Acceptance:

- Local combat resolves and persists casualties/loot.
- Attack limits are enforced.
- Defenders receive dispatches.
- Ballistic and spy actions consume required resources.
- Combat tests cover win, loss, and limit behavior.

## D4 - InterDOOR Integration

Goal: Connect Empire Ascendant to InterDOOR federation features.

Status: D4A foundation, D4B cross-node PvP, D4C Hyperdrive travel foundation, D4D inactive empire lifecycle, and D4E local SSH listener foundation are implemented and accepted. Deployment, live public listing, full visitor gameplay, destructive purge tooling, and polish remain deferred.

Slices:

1. Register and heartbeat with hub.
2. Emit Empire Ascendant events.
3. Pull and display roster.
4. Add cross-node rankings.
5. Drain cross-node attack queue through existing sync patterns.
6. Implement cross-node ground attacks.
7. Implement cross-node ballistic strikes.
8. Implement Galactic News feed.
9. Implement Hyperdrive travel state.
10. Implement visitor restrictions while away from home node.
11. Implement inactive empire warning, hiding, and purge.
12. Implement local SSH listener foundation for local review access.
13. Add tests or controlled smoke tests for federation and SSH behavior.

Acceptance:

- Local-only play still works when federation is unavailable.
- Federation failures degrade cleanly.
- Cross-node actions are queued and resolved by the proper node.
- Travel state is persisted and reversible.
- Local SSH listener can run the existing terminal session without changing stdio mode.

## D5 - Polish And Balance

Goal: Improve presentation, instructions, lore, and balance without changing the architecture.

Slices:

1. Add ANSI title art or adapt existing ANSI assets. Accepted in D5B.
2. Add story/instructions screen. Accepted in D5A and revised in D5B.
3. Improve menu copy and terminal presentation. Accepted in D5A/D5B.
4. Add leaderboard scoring review. Next D5C work.
5. Simulate representative empire states. Next D5C work.
6. Revisit money score coefficient. Next D5C work.
7. Balance production, combat ratios, and action costs. Next D5C work.
8. Document final balance decisions. Next D5C work.

Acceptance:

- Terminal UI is coherent and readable.
- Instructions explain core play without a web/manual dependency.
- Balance changes are backed by tests, simulations, or documented reasoning.
- D5 does not introduce unnecessary architecture changes.

## Deployment Phase

Goal: Deploy only after local D1-D5 scope is ready enough for public testing.

Slices:

1. Inspect authoritative infrastructure documentation.
2. Confirm host, ports, service names, binary path, and DB path.
3. Build release binary using existing release process.
4. Install binary.
5. Create systemd unit.
6. Create data directory with appropriate ownership.
7. Open firewall ports only after explicit approval.
8. Start service.
9. Verify local and remote connection.
10. Update infrastructure documentation.

Acceptance:

- Service starts under systemd.
- Database path is correct.
- Port is reachable as intended.
- Rollback path is known.
- Infrastructure documentation is updated.
