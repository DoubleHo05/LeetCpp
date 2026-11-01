# AI-LOG.md

Chronological, dated log of AI-assisted work on this repository (LeeTKOB /
LeetCpp). Entries record what was decided, what actually went wrong, and how
it was fixed. Rules in `AGENTS.md` that grew out of an entry cite the entry
number so the rule can be traced back to the incident.

---

## 2026-10-31 #1 — PA#2 harness: contract first, then code

**What:** Started the PA#2 harness/quality-gate milestone. Chose
`POST /labs/{labId}/submissions` as the endpoint to contract, per the brief's
guidance ("the endpoint whose errors you would least like a stranger to see").
A retried upload silently creating a second submission is exactly that kind of
error.

**Decision:** Wrote `docs/contract/post-labs-labid-submissions.md` before any
handler code. The contract is the specification; the implementation in
`backend/internal/api/submit.go` and the tests in `submit_test.go` serve it.
When code and document disagree, the document wins.

## 2026-10-31 #2 — Incident: test harness hard-coded expectations

**What went wrong:** The first `go test` run of the contract suite failed 2
tests. Both failures were bugs in the **test helpers**, not in the handler:

1. `wantCreated` hard-coded `studentId == "u-alice"`, which is wrong for the
   test that proves idempotency keys are scoped per student (`u-sam` submits
   too).
2. The request helper auto-filled `Idempotency-Key`, so the "empty key"
   request silently became a valid 201 instead of exercising the
   `400 invalid-idempotency-key` rule.

**Fix:** Corrected the helpers so they express the contract (expected student
id is now a parameter; the empty-key case builds its own request). No
assertion was relaxed; the handler was not touched.

**Rule that grew from it:** `AGENTS.md` → Never → "Never modify tests just to
make implementation pass." The distinction: a broken *harness* may be fixed to
match the contract; a failing *assertion* means the code (or, after a contract
change, the contract first) is what gets fixed.

## 2026-10-31 #3 — Incident: the format check caught a new file

**What went wrong:** The first run of the quality gate reported
`backend/internal/api/submit_test.go` from `gofmt -l` — a newly written file
was not formatted.

**Fix:** `gofmt -w backend`. The gate is green again.

**Rule that grew from it:** `AGENTS.md` → Never/conventions → run `make check`
before every push; never merge on a red `make format-check`.

## 2026-10-31 #4 — Deliberate contract break (evidence that the tests bite)

**What:** To demonstrate that the contract tests actually catch violations,
the handler was broken on purpose: `idempotency-key-reused` returned `400`
instead of the contracted `409`.

**Result:** `go test` went red:

```
submit_test.go:274: status = 400, want 409
--- FAIL: TestIdempotencyKeyReused409
```

**Fix:** Restored `http.StatusConflict` in `submit.go`. The test file was not
edited in the fix — the code violated the contract, the test was right.

**Evidence plan:** once the team authorizes commits/push, this becomes a
break-commit with CI red, then a fix-commit with CI green, and the links go
into `README.md` §8 as the "gate bites" evidence required by the rubric.

## 2026-10-31 #5 — Decision: idempotent replay resolves before business rules

**What:** Error-precedence design question: a retry of an already-accepted
upload arrives after the submission window has closed. Return the original
result, or reject with `submission-window-closed`?

**Decision:** The contract's precedence list (§5) resolves the idempotent
replay *before* lab/role/window rules. A retry never creates, and "safe to
retry" must hold at exactly the wrong moment too. Covered by
`TestReplayWinsOverClosedWindow`.

**Rule that grew from it:** `AGENTS.md` → Never → never reorder the
error-precedence list in code alone; contract, implementation, and tests
change together.

## 2026-10-31 #6 — Decision: spending cap before any model key exists

**Decision:** README §7 now states the model key carries a hard cap of
**USD 15 per calendar month**, set in the provider's billing console before
the key is issued. No provider is enabled yet, so no key exists yet — the cap
rule is in place first. Unit tests must never call the real model
(`AGENTS.md` → Never).

## 2026-10-31 #7 — Decision: CI runs `make check` verbatim

**Decision:** `.github/workflows/ci.yml` runs exactly `make check` — the same
command the README documents. No CI-only scripts, so "works on my machine"
and CI cannot diverge. Branch protection on `main` must require the `checks`
status (README §8) so a red check blocks the merge instead of merely
complaining.

## 2026-10-31 #8 — Secrets and history audit

**Checked:** `git log` contains a single commit (`267b579 Initial commit`,
README only) — no secret has ever been committed. `.gitignore` covers
`.env`/`.env.*` with an exception for `.env.example`, which lists variable
names without values. The submission sandbox must never receive the model key
or any application secret (`AGENTS.md` → Never).
