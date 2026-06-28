# Testing And Verification

This file records the verification approach for Empire Ascendant work.

## General Rule

Every implementation session should end with repeatable verification commands and a short result summary.

For local LLM work, require the model to report:

- commands run
- whether each command passed
- relevant failure output
- any tests not run and why

## Before InterDOOR Repo Inspection

The local D1 commands are established for this folder. Do not inherit commands from another game unless they are copied and adapted locally on purpose.

Local commands:

```bash
make help
make test
make build
make run
make smoke
go test ./...
```

## D1 Verification Targets

D1 is not complete until these are verified somehow:

- schema creation or migration for `empires`
- player creation persists World Name and Empire Name
- duplicate World Name is rejected
- duplicate Empire Name is rejected
- daily turn reset uses Unix day number logic
- Empire Report displays stored D1 resources
- binary builds
- basic session path reaches main menu and Empire HQ shell

## Manual Smoke Test Template

Use this template after a D1 build exists:

```bash
# Build the Dominion binary using the repository's normal build command.

# Start the Dominion node locally with a temporary SQLite database.

# Connect through the existing local terminal/SSH path.

# Create a new player.

# Confirm:
# - world name appears correctly
# - empire name appears correctly
# - turns show 15
# - report displays money, population, food, energy, and starter resources
# - quitting exits cleanly
```

Fill in exact commands after the folder-local build/test path has been created.

## Deployment Verification

Deployment is not part of D1. Before any deployment work:

- inspect the infrastructure source of truth before any host/service/domain change
- confirm host, ports, systemd unit, database path, and firewall rules
- review impact before making changes
- update infrastructure documentation if host/service/domain state changes


## Verified On 2026-06-24

Commands run successfully:

```bash
go test ./...
make build
./bin/interdoor-dominion -db ':memory:' -stdio
```

Manual smoke path verified:

- main menu appears
- new empire creation works
- World Name and Empire Name persist in the session
- Empire HQ appears with 15 turns
- Empire Report shows money 10000, population 20000, food 10000, energy 300
- quit to main and quit program work

## D2 Verified On 2026-06-24

Commands run successfully:

```bash
go test ./...
make build
make smoke
./bin/interdoor-dominion -db ':memory:' -stdio
```

Manual smoke path verified:

- new empire creation works
- Develop Empire menu opens
- activating an Agricultural region costs one turn
- bank deposit does not cost a turn
- Empire Report shows updated money, bank, turns, regions, buildings, mines, and energy tech
- quit to main and quit program work

## D2 Legacy Backfill Verified On 2026-06-24

Commands run successfully:

```bash
go test ./...
make build
./bin/interdoor-dominion -db var/empireascendant.db -stdio
```

Existing D1-created empire records without D2 economy rows now backfill economy state on login.

## D3 Verified On 2026-06-24

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nA\n1\n1\n5\nQ\nT\nQ\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d3-smoke.db -stdio
```

Automated tests verified:

- military rows and D3 tech keys are seeded for new empires
- D3 military rows backfill for existing empires without resetting researched tech
- military and D3 tech changes persist through `SaveEconomy`
- ground attack win and loss behavior
- recruitment, missile, and spy actions consume turns, money, missiles, or spies as appropriate
- per-target/day attack limits are enforced
- defender dispatches can be created and listed
- Attack Menu recruitment works through the terminal session path

Manual smoke path verified:

- new empire creation works
- Attack Menu opens from Empire HQ
- recruiting normal soldiers costs one turn and money
- Empire Report shows updated turns, money, soldiers, defenses, weapons, and spies
- quit to main and quit program work

## D4A Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nW\nQ\nR\nN\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d4a-smoke.db -node-id ascendant -stdio
```

Automated tests verified:

- InterDoor client sends registration, heartbeat, event, and roster requests with expected JSON shapes
- structured hub errors are surfaced clearly
- local federation event log allocates monotonic source-node event IDs
- pushed events are marked so they are not resent
- pulled events are stored idempotently and projected into Galactic News
- local roster entries use Empire Ascendant score as the public level field
- remote roster entries are stored and stale-aware
- sync-once fake-hub flow performs heartbeat, event push/pull, roster push/pull, and cursor persistence
- terminal session displays Rankings, Galactic News, and Wanderers

Manual smoke path verified:

- local stdio mode still works without live hub access
- new empire creation emits local news when `-node-id` is supplied
- Wanderers displays an empty remote state without error
- Rankings shows the local empire score
- Galactic News shows local event-derived news

Live hub registration was not run during D4A verification. D4A tests use fake hubs only.

## D4B Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nW\nQ\nR\nN\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d4b-smoke.db -node-id ascendant -stdio
```

Automated tests verified:

- InterDoor client sends PvP queue, pending, result, and blocked requests with expected JSON shapes
- outbound PvP queue tracking records hub request IDs for local attackers
- victim-side pending PvP resolution mutates only local victim state and emits `pvp.resolved`
- repeated inbound pending requests are idempotent
- pulled `pvp.resolved` events apply attacker-side casualties and ground loot exactly once
- sync-once fake-hub flow drains pending PvP, completes resolved requests, blocks malformed payloads, and pushes result events
- terminal session can queue a remote ground attack from a remote roster target
- local stdio play still works without live hub access

Manual smoke path verified:

- new empire creation works
- Wanderers, Rankings, and Galactic News still render after D4B wiring
- local stdio mode exits cleanly

Live hub registration, live PvP queueing, travel, SSH listener, deployment, and polish were not run during D4B verification.

## D4C Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nW\nQ\nR\nN\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d4c-smoke.db -node-id ascendant -stdio
```

Automated tests verified:

- InterDoor client sends travel submit, pending, arrived, and blocked requests with expected JSON shapes
- local empire travel snapshots export with global ID, home node, and full economy state
- accepted travel marks the local empire as traveling
- destination nodes import pending travel arrivals idempotently
- malformed travel snapshots are rejected for blocking
- sync-once fake-hub flow imports arrivals, marks them arrived, blocks malformed arrivals, and pushes `player.traveled`
- terminal session can submit Hyperdrive travel to a known remote node
- home-node login is blocked while the empire is traveling away
- local stdio play still works without live hub access

Manual smoke path verified:

- new empire creation works
- Empire HQ includes Hyperdrive
- Wanderers, Rankings, and Galactic News still render after D4C wiring
- local stdio mode exits cleanly

Live hub registration, live travel queueing, SSH listener, deployment, full visitor gameplay, inactive empire lifecycle, and polish were not run during D4C verification.

## D4C Review Fix Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
```

Automated tests verified:

- username lookup is case-insensitive
- duplicate username creation is blocked across casing differences
- terminal login with different username casing does not offer to create a second empire

Existing databases receive a `username_key` backfill during store initialization.

## D4D Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nW\nQ\nR\nN\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d4d-smoke.db -node-id ascendant -stdio
```

Automated tests verified:

- lifecycle transitions from active to warned, hidden, and purge-eligible
- purge-eligible status is non-destructive
- login reactivates hidden empires
- hidden empires are omitted from rankings, roster push, heartbeat player count, and local attack targets
- terminal login shows a reactivation notice for hidden empires
- normal local create/login/menu flow still works

SSH listener, deployment, live hub registration, destructive purge tooling, full visitor gameplay, and polish were not run during D4D verification.

## D4E Verified On 2026-06-25

Verification commands:

```bash
go test ./...
make build
make smoke
timeout 2s ./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:0 -db /tmp/empireascendant-d4e-cli.db -ssh-host-key /tmp/empireascendant-d4e-ssh-host-key
```

The listener smoke command exits with status `124` because `timeout` stops the long-running SSH listener after startup. Expected output includes `Empire Ascendant SSH listener on 127.0.0.1:0`.

Automated tests verified:

- missing SSH host keys are generated with `0600` permissions
- existing SSH host keys are reused
- loopback SSH clients can authenticate with a password, request PTY/shell, send `Q`, and receive session output
- SSH terminal input echoes typed printable characters, handles Backspace, and normalizes OpenSSH Enter keys for existing line-based prompts
- stdio mode and the existing build path still work

D4E implementation adds `golang.org/x/crypto/ssh` for local SSH transport.

Manual SSH review command:

```bash
./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:2324 -db /tmp/empireascendant-d4e-ssh.db
ssh -p 2324 review@127.0.0.1
```

The `/tmp/empireascendant-d4e-ssh.db` path intentionally starts with a fresh review database. Use `-db var/empireascendant.db` or another known existing database path when reviewing existing empires.

Any SSH password is accepted in D4E local review mode. The game still prompts for the Empire Ascendant username/password after the SSH connection opens.

Public deployment, firewall/systemd changes, live hub listing, live registration token use, full visitor gameplay, destructive purge tooling, and polish were not run during D4E verification.

D4E user review confirmed SSH connection, terminal input, and game session access. The temporary review database caveat was documented: `/tmp/empireascendant-d4e-ssh.db` does not contain empires created in another database path.

## D5A Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'S\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5a-help.db -stdio
```

Automated tests verified:

- Story / Instructions renders from the main menu
- the old `Not available yet.` Story stub is gone
- presentation helpers print section headers and menu rows
- existing session flows still pass

Manual smoke path verified:

- main menu renders with an ASCII Empire Ascendant frame
- `S` displays Story, How To Play, and Current Limits sections
- the screen returns to the main menu after Story / Instructions
- `Q` exits cleanly

D5A did not change mechanics, balance constants, database schema, federation behavior, SSH transport behavior, deployment behavior, full visitor gameplay, or purge tooling.

## D5B Verified On 2026-06-25

Commands run successfully:

```bash
go test ./...
make build
make smoke
printf 'S\nQ\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-help.db -stdio -ansi=false
printf 'Q\n' | ./bin/interdoor-dominion -db /tmp/empireascendant-d5b-revised-ansi.db -stdio -ansi=true
timeout 2s ./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:0 -db var/empireascendant.db -ansi=true
```

The listener smoke command exits with status `124` because `timeout` stops the long-running SSH listener after startup. Expected output includes `Empire Ascendant SSH listener on 127.0.0.1:0`.

Automated tests verified:

- plain presenter output includes expected title/menu text and no ANSI escapes
- ANSI presenter output includes escape sequences, command text, and BBS-style frame characters
- existing session flows still pass

Manual smoke path verified:

- `-ansi=false` shows readable Empire Ascendant title art and menu panels without escape codes
- `-ansi=true` emits compact 80-column ANSI BBS-style title/menu output with black-background panels and bright command keys
- existing `var/empireascendant.db` initializes and reaches ANSI-enabled listener startup

D5B did not change mechanics, balance constants, database schema, federation behavior, SSH transport behavior, deployment behavior, full visitor gameplay, or purge tooling.

## D5B Final Acceptance Verified On 2026-06-28

Commands run successfully during final D5B review and cleanup:

```bash
go test ./...
make build
make smoke
printf 'Q\n' | ./bin/interdoor-dominion -stdio=true -db /tmp/empireascendant-plain-smoke.db -ansi=false
printf 'Q\n' | ./bin/interdoor-dominion -stdio=true -db /tmp/empireascendant-ansi-smoke.db -ansi=true
./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:2324 -db var/empireascendant.db -ssh-host-key var/ssh_host_rsa -ansi=true -ssh-encoding=auto
```

Manual review verified:

- ANSI visuals and aesthetics look good in SyncTerm.
- `-ssh-encoding=auto` renders correctly in both SyncTerm and a regular terminal SSH session.
- Banking deposit/withdraw flow works and updates visible financial status.
- Bank invalid commands stay in Bank and show an error.
- Develop/worker/attack blocked actions show visible feedback.
- Army Status uses readable color-separated sections.
- Plain fallback sanity checks passed.
- The standalone GitHub repository was created at `njb1966/empireascendant`.
- Repository identity cleanup removed old reference screenshots/sample ANSI screens and replaced the historical README with an Empire Ascendant README using `screenshot.png`.

D5B is accepted. The next D5 slice is D5C: leaderboard scoring review, representative empire simulations, money score coefficient review, production/combat/action-cost balance review, and final balance decision documentation.

## Legacy Username Migration Fix Verified On 2026-06-25

Commands run successfully:

```bash
go test ./internal/data
go test ./...
make build
make smoke
timeout 2s ./bin/interdoor-dominion -stdio=false -addr 127.0.0.1:0 -db var/empireascendant.db
```

The listener smoke command exits with status `124` because `timeout` stops the long-running SSH listener after startup. Expected output includes `Empire Ascendant SSH listener on 127.0.0.1:0`.

Verified fix:

- legacy databases without `empires.username_key` now add/backfill the column before creating `idx_empires_username_key`
- SQLite store uses a single open connection to avoid DDL lock contention
- username backfill closes its scan before applying updates
- existing `var/empireascendant.db` initializes and reaches listener startup

## D5C Scoring And Balance Review Verified On 2026-06-28

Commands run successfully:

```bash
go test ./internal/game
go test ./...
make smoke
```

Automated tests verified:

- score formula includes population, military attack power, total wealth, and unlocked tech
- free banking does not change score
- money score coefficient is effectively `0.1` through `ScoreMoneyDivisor = 10`
- representative starting empire score is `23000`
- representative economic builder score is `83000`
- representative raider score is `44500`
- early mining path is viable: activate Agricultural, accrue three production ticks, build Miners Guild, hire/assign a miner, mine and sell first gold output
- default combat balance holds: modest recruitment loses against default defenses; committed recruitment can win but remains net-costly against a default treasury

D5C changed leaderboard scoring only. It did not change production constants, combat strengths, action costs, database schema, federation behavior, SSH transport behavior, deployment behavior, full visitor gameplay, or purge tooling.
