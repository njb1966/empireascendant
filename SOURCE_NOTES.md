# Source Notes

This file captures project facts discovered from the original Dominion files. Use it when a task involves original defaults, source-model translation, or conflicts between historical files.

## Source Files

- `DOMINION.PAS`: main Pascal source. Contains data records, config reader copy, player creation, menus, and an empty `Play_Game`.
- `SETUP.PAS`: setup/config parsing source.
- `A.PAS`: assignment list for config values.
- `CONFIG.DOM`: original runtime configuration file.
- `MENU*.ANS` and `SCREENS/*.ANS`: ANSI art/menu assets.
- `README.md`: historical background from the original author.

## Critical Fact

`DOMINION.PAS` has a declared `Play_Game` procedure, but the implementation is empty:

```pascal
Procedure Play_Game;
begin
end;
```

That means the rewrite must infer gameplay from the data model, config, menus, BBS genre context, and `REWRITE_PLAN.md`. It must not claim to port implemented gameplay.

## Startup Defaults From CONFIG.DOM

Use these when original startup defaults matter:

| Field | Value |
|---|---:|
| daily turns | 15 |
| money | 10000 |
| money in bank | 0 |
| population | 20000 |
| food | 10000 |
| energy | 300 |
| normal soldiers | 100 |
| recon level 1 drones | 5 |
| defense satellites | 10 |
| ground turrets | 10 |
| fossil fuel tech | true |
| normal human soldier tech | true |
| level 1 drone tech | true |
| miners guild | false |
| miners | 0 |
| fishing guild | false |
| fishers | 0 |
| fish stock | 0 |
| construction factories | 1 |
| research labs | 2 |
| intelligence building | false |
| spies | 0 |
| terrorists | 0 |
| lottery | false |
| gold mines | 1 |
| silver mines | 1 |
| iron mines | 1 |
| copper mines | 0 |
| nickel mines | 0 |
| agricultural regions | 10 |
| industrial regions | 1 |
| desert regions | 0 |
| urban regions | 0 |
| river regions | 0 |
| ocean/sea regions | 0 |
| wastelands | 0 |

All other listed soldiers, weapons, vehicles, advanced defenses, and advanced techs default to zero or false unless noted above.

## Known Source Conflict

The top comment in `DOMINION.PAS` says new players start with values such as `$20,000`, `27,426` population, `1` building point, `1` research point, and `3` industrial regions. `CONFIG.DOM` says otherwise.

Decision: `CONFIG.DOM` wins for original defaults because it is the parsed runtime config file.

## Pascal Data Model Guidance

The `One_Player` record is the primary source for original field names and domain concepts. Modern Go/SQLite code should use clear Go naming while preserving the concepts:

- world name
- empire name
- daily turns
- money and banked money
- food and food storage
- energy
- population
- regions
- mines
- tech flags
- military units
- guilds/buildings
- defenses
- messages/news

Do not blindly recreate every Pascal field in D1. Follow the phase scope in `PHASE_PLAN.md`.
