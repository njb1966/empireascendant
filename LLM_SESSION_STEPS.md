# LLM Session Steps

This is the short operator checklist for working with `qwen3-coder:30b` through `codepal`.

## 1. Start With The Control Files

Use this for normal implementation sessions:

```bash
codepal -i PROJECT.md \
        -i DECISIONS.md \
        -i CURRENT_TASK.md \
        -i SESSION_HANDOFF.md \
        -i D5B_VISUAL_HANDOFF.md \
        -i REWRITE_PLAN.md \
        -i INTERDOOR_DESIGN_GUIDE.md
```

Use this when the task involves original Pascal/config details:

```bash
codepal -i PROJECT.md \
        -i DECISIONS.md \
        -i CURRENT_TASK.md \
        -i SESSION_HANDOFF.md \
        -i D5B_VISUAL_HANDOFF.md \
        -i REWRITE_PLAN.md \
        -i INTERDOOR_DESIGN_GUIDE.md \
        -i SOURCE_NOTES.md \
        -i CONFIG.DOM \
        -i DOMINION.PAS \
        -i SETUP.PAS
```

## 2. Ground The Model

Paste:

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

Correct anything wrong before continuing.

## 3. Ask For A Standup

Paste:

```text
Standup:
- Summarize what is done.
- Summarize what is next.
- Identify blockers or unknowns.
- Identify the smallest useful slice for this session.
- Do not edit files yet.
```

## 4. Ask For A File-Level Plan

Paste:

```text
Plan the smallest useful slice for this session.

Before editing, provide:
- source evidence by file,
- exact local files you intend to inspect,
- exact local files you expect to edit,
- acceptance criteria,
- verification commands,
- risks or stop conditions.

Do not implement until I approve the plan.
```

Approve only if the plan is phase-scoped and source-backed.

## 5. Let It Implement

Paste:

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

## 6. Require Verification

Paste:

```text
Run the agreed verification commands.

Report:
- command,
- pass/fail,
- relevant output,
- any verification not run and why.
```

## 7. Require Diff Review

Paste:

```text
Review your own changes against the acceptance criteria.

Report:
- files changed,
- behavior changed,
- risks,
- what remains,
- whether this slice is complete.
```

## 8. Update The Project Memory

Paste:

```text
Update only the project memory needed for this completed slice:
- PROJECT.md Done and Next Task,
- CURRENT_TASK.md if the active slice changes,
- DECISIONS.md for settled decisions,
- SOURCE_NOTES.md for new source findings,
- TESTING.md for verified commands.

Keep updates concise.
```

## 9. End The Session

After a completed slice:

```text
Summarize the final state in 10 lines or fewer.
```

Then quit or clear the session. Start fresh for the next slice.

## Drift Recovery

If the model drifts, paste:

```text
Stop. Re-read PROJECT.md source precedence and Do Not list.
State the current phase and exact task again.
Do not continue implementation until corrected.
```

If it drifts again, clear or restart the session and reload the control files.
