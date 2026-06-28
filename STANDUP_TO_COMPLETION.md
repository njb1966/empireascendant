# Standup To Completion Workflow

Use this workflow for each local LLM session with `codepal` and `qwen3-coder:30b`.

## Purpose

The model has no durable memory and can drift during long sessions. The project documents are the memory. Each session should load the documents, confirm the current task, do one scoped slice, verify it, and update the documents.

## Folder Boundary

Run sessions from this folder. Keep work inside this folder and subfolders. Outside-folder files may be read only to confirm current InterDoor protocol/interface requirements, and they must not become design authority for Empire Ascendant.

## Session Start

Run from the project or implementation workspace:

```bash
codepal -i PROJECT.md \
        -i DECISIONS.md \
        -i CURRENT_TASK.md \
        -i REWRITE_PLAN.md \
        -i INTERDOOR_DESIGN_GUIDE.md
```

Add task-specific files only when needed:

```bash
codepal -i PROJECT.md \
        -i DECISIONS.md \
        -i CURRENT_TASK.md \
        -i REWRITE_PLAN.md \
        -i INTERDOOR_DESIGN_GUIDE.md \
        -i SOURCE_NOTES.md \
        -i CONFIG.DOM \
        -i DOMINION.PAS \
        -i SETUP.PAS
```

Use the larger command only for source archaeology or default-value translation.

## Grounding Prompt

```text
Read the loaded project documents. Do not write code yet.

Confirm:
1. what this project is,
2. the current phase,
3. the exact current task,
4. hard constraints,
5. source precedence,
6. local files you expect to inspect,
7. assumptions that need source evidence.

If any instruction conflicts, identify the conflict instead of resolving it silently.
```

Correct any misunderstanding before giving a task.

## Standup Prompt

```text
Standup:
- Summarize what is done.
- Summarize what is next.
- Identify blockers or unknowns.
- Identify the smallest useful slice for this session.
- Do not edit files yet.
```

## Planning Prompt

```text
For this session, plan only the current slice.

Before editing, provide:
- source evidence by file,
- exact local files you intend to inspect,
- exact local files you expect to edit,
- acceptance criteria,
- verification commands,
- risks or stop conditions.

Do not implement until I approve the plan.
```

## Implementation Prompt

```text
Proceed with the approved slice.

Constraints:
- keep edits minimal and phase-scoped,
- follow local Empire Ascendant project patterns,
- do not add dependencies without asking,
- do not implement later-phase features,
- update or add focused tests where appropriate,
- report any source conflict instead of guessing.
```

## Review Prompt

```text
Review your own diff against the acceptance criteria.

Report:
- files changed,
- behavior changed,
- commands run,
- test results,
- risks,
- documentation updates needed.
```

## Completion Prompt

```text
Update the project docs for this completed slice:
- PROJECT.md Done and Next Task,
- CURRENT_TASK.md if the active slice changes,
- DECISIONS.md for settled decisions,
- SOURCE_NOTES.md for new source findings,
- TESTING.md for verified commands.

Keep updates concise.
```

## When To Clear Or Restart

Clear the session or restart `codepal` when:

- the model stops respecting source precedence
- it proposes unrelated rewrites
- it invents InterDOOR APIs
- it starts implementing later phases
- it repeats a corrected mistake
- the session has grown long enough that summaries replace source evidence

After clearing, reload the control files and use the grounding prompt again.

## Completion Standard

A slice is complete only when:

- acceptance criteria are met
- verification commands have been run or explicitly blocked
- the diff has been reviewed
- project docs reflect the new state
- the next task is clear
