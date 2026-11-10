# AI-LOG.md

Chronological, dated log of AI-assisted work on this repository (LeeTKOB /
LeetCpp). Entries record what was decided, what actually went wrong, and how
it was fixed. Rules in `AGENTS.md` that grew out of an entry cite the entry
number so the rule can be traced back to the incident.

## Tools used, and for what

| Tool | Part of the work |
|---|---|
| **MiMoCode** (CLI coding assistant) | All PA#2 work: the contract draft, the Go handler and contract tests, `ci.yml`, `AGENTS.md`, the PA#2 evidence audit, the deliberate contract break, the branch-protection fix, `PA2.md`, `SELF_ASSESSMENT_REPORT.md` |
| Go toolchain (`go test`, `go vet`, `gofmt`) | Backend checks and the contract suite |
| pnpm + eslint + prettier + tsc + vitest | Frontend checks |
| GitHub Actions, `checks` job | Runs `make check` on every PR and on the `pa2` tag |
| GitHub REST API | Reading and correcting the `main` branch-protection rule |

Entries #1–#9 come from the initial PA#2 harness build; entries #10 onward
also record the tool where it matters to the incident.

## Changed / Rejected

| | Entry | What |
|---|---|---|
| **Changed** | #11 | The CI workflow the assistant wrote could never run for the `pa2` tag the rubric asks for — found while filling in `PA2.md` links. |
| **Changed** | #12 | The branch-protection rule required *no* status check, so a red PR was still mergeable — found by reading the rule over the API instead of trusting the README. |
| **Changed** | #2, #9 | Test helpers that hard-coded expectations; a `typecheck` script that depended on a local `.next/` cache. |
| **Rejected** | #13 | Force-moving the `pa2` tag so the tag's own CI-run link would land inside the tagged tree. |

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

## 2026-11-01 #9 — Incident: CI caught a "works on my machine" typecheck

**What went wrong:** The first real CI run of `make check` went **red**:
`app/layout.tsx(20,50): error TS2304: Cannot find name 'LayoutProps'` — while
`pnpm typecheck` had been green locally the whole time.

**Root cause:** `LayoutProps` is a *generated* global type (Next 16 writes it
into `.next/types/routes.d.ts` during `next dev` / `next build` /
`next typegen`). A local `.next/` cache from earlier runs made `tsc` pass; CI
starts from a fresh checkout with no `.next/`, so the type did not exist. Same
command, different machine state — exactly the failure the "same commands
locally and in CI" rule exists to expose.

**Fix:** the frontend `typecheck` script is now `next typegen && tsc --noEmit`
(per Next 16's docs: these helpers are "generated when running either
`next dev`, `next build` or `next typegen`"). Verified by deleting `.next/`
and running the check clean — green. `make check` green end to end.

**Rule reinforced:** a local green that depends on build artifacts is not
evidence; only a fresh-clone run is. The red CI run on the PA#2 PR was the
gate doing exactly its job.

## 2026-11-11 #10 — The deliberate contract break, executed for real

**What:** Entry #4 above rehearsed the contract break locally and left an
"evidence plan" for later. That plan is now done. Tool: MiMoCode.

**What happened, in order:**

1. `backend/internal/api/submit.go` was changed so `idempotency-key-reused`
   returned `400` where
   [`docs/contract/post-labs-labid-submissions.md`](docs/contract/post-labs-labid-submissions.md)
   §5 contracts `409`. `submit_test.go` was not opened or edited.
2. Commit `6805acb` pushed to `pa2/contract-break`, PR #2 opened. `checks`
   went **red** — run
   [`42879941393`](https://github.com/DoubleHo05/LeetCpp/actions/runs/42879941393):
   `submit_test.go:274: status = 400, want 409`.
3. Merging was attempted while red and GitHub refused it outright:
   `Required status check "checks" is failing.` The PR reported
   `mergeable_state: blocked`.
4. Commit `9047e6e` restored `http.StatusConflict`. `checks` went **green** —
   run [`42880430692`](https://github.com/DoubleHo05/LeetCpp/actions/runs/42880430692).
5. PR #2 merged with a merge commit, so both commits stay in `main`'s history
   as the evidence.

**Changed:** nothing in the test suite. The only file touched in either commit
is `submit.go`. That is the point of the exercise: the code violated the
contract, the assertion was right.

## 2026-11-11 #11 — Changed: the CI workflow could never run for the tag

**What went wrong:** The assistant wrote `.github/workflows/ci.yml` (entry #7)
triggering on `pull_request` and on `push` to `main` only. PA#2 asks for "the
CI run for the tag", and pushing the `pa2` tag starts **no** workflow. The
evidence link in `PA2.md` could never be filled in — a check that cannot run
is the same defect as a check that cannot fail.

**How it was found:** not by reading the workflow, but by trying to fill in
`PA2.md` and realising the slot had no possible value. Writing the evidence
first exposes the gaps that reading the code does not.

**Fix:** `on: push: tags: [pa2]` added to the same workflow. Same job, same
command, still `make check`.

**Rule reinforced:** every artifact the rubric asks for must have a real way
to come into existence. If it cannot, the harness is wrong.

## 2026-11-11 #12 — Changed: branch protection required no status check

**What went wrong:** `main` *had* a protection rule (PR required, admins
enforced, force pushes off), and `README.md` §8 said the `checks` status must
be required. The rule's `required_status_checks` was in fact **empty**. A red
PR could have been merged. The documentation and the setting disagreed.

**How it was found:** the assistant read the rule over the GitHub API rather
than trusting the README's description of it — `GET /branches/main/protection`
returned `contexts: []`. Documentation of a safety setting is not evidence
that the setting exists.

**Fix:** `checks` is now a required status on `main`, with `strict: true`
(the branch must be up to date and the check green before merging). The API
response after the change is kept as
[`docs/evidence/branch-protection.json`](docs/evidence/branch-protection.json).
The refusal quoted in entry #10 step 3 is the same setting biting.

**Rule reinforced:** never document a safety control without reading the
control back. `AGENTS.md` → Never → "never weaken sandbox restrictions or CI
checks" covers the failure mode; this entry is the read-back habit that
catches the silent version of it.

## 2026-11-11 #13 — Rejected: force-moving the `pa2` tag

**What was proposed:** PA#2 wants "the CI run for the tag" linked from
`PA2.md`, but the run URL only exists after the tag is pushed — so the tagged
tree cannot contain the link to its own run. The assistant considered
finishing `PA2.md`, tagging, then `git tag -f pa2` to re-point the tag at a
follow-up commit that records the URL.

**Rejected.** Force-moving a published tag is exactly the kind of destructive,
hard-to-reverse git action `AGENTS.md` forbids, and it would make the tag's
tree lie about its own history. Instead: the tag points at the deliverables
commit, `PA2.md` links the `checks` run of that exact commit (the run every
reviewer sees next to it), and the tag's own run URL is recorded in the
submission zip, which is taken from `main` after that commit. The evidence is
complete without rewriting anything.
