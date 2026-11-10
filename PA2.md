# PA#2 — Harness and quality gate · submission links

Product: **LeeTKOB** (repository `LeetCpp`). Course: CSC13114, PA#2.
This file is the evidence index the brief asks for.

| # | What | Link | Status |
|---|---|---|---|
| 1 | Repository | <https://github.com/DoubleHo05/LeetCpp> | public |
| 2 | Protected-branch settings | [`docs/evidence/branch-protection.json`](docs/evidence/branch-protection.json) | done |
| 3 | Blocked PR (red check refused) | <https://github.com/DoubleHo05/LeetCpp/pull/2> | done |
| 4 | PR that turned green | <https://github.com/DoubleHo05/LeetCpp/pull/2> (same PR) | done |
| 5 | CI run for the tag `pa2` | see [§5](#5-ci-run-for-the-tag-pa2) | recorded at tag time |
| 6 | Deliberate contract break | <https://github.com/DoubleHo05/LeetCpp/pull/2> | done |

---

## 1. Repository

<https://github.com/DoubleHo05/LeetCpp> — public, default branch `main`.

## 2. Protected-branch settings

Evidence: [`docs/evidence/branch-protection.json`](docs/evidence/branch-protection.json)
— the GitHub REST response for
`GET /repos/DoubleHo05/LeetCpp/branches/main/protection`, saved verbatim after
the rule was corrected. It shows:

| Setting | Value |
|---|---|
| `required_status_checks.contexts` | `["checks"]` |
| `required_status_checks.strict` | `true` |
| `required_pull_request_reviews` | present (a PR is required before merging) |
| `enforce_admins.enabled` | `true` — no admin bypass |
| `allow_force_pushes.enabled` | `false` |
| `allow_deletions.enabled` | `false` |

The rule was **corrected** as part of this submission: before the fix,
`required_status_checks` was empty, so a red PR was still mergeable. The
incident is recorded in [`AI-LOG.md`](AI-LOG.md) 2026-11-11 #12. A browser
screenshot of the same Settings page is optional extra evidence; the API
response is the machine-readable truth of the rule.

## 3. Blocked PR (red check refused the merge)

<https://github.com/DoubleHo05/LeetCpp/pull/2> — "PA#2 evidence: deliberate
contract break (expected red, then fix)".

Commit `6805acb` violated the contract (see §6). Merging it was attempted
while red and **GitHub refused**:

```
Required status check "checks" is failing.
```

The PR reported `mergeable_state: blocked`. Red run:
[`actions/runs/42879941393`](https://github.com/DoubleHo05/LeetCpp/actions/runs/42879941393)
(`checks` → `failure`, `submit_test.go:274: status = 400, want 409`).

An earlier red→green pair also exists on
[PR #1](https://github.com/DoubleHo05/LeetCpp/pull/1) (the PA#2 harness work):
`b0bb518` went red on a fresh-clone typecheck
([run 41311559092](https://github.com/DoubleHo05/LeetCpp/actions/runs/41311559092))
and the PR was not merged until `bb94186` turned it green
([run 41312466512](https://github.com/DoubleHo05/LeetCpp/actions/runs/41312466512)).
Note that at the time of PR #1 the protection rule had no required status
check, so that red was respected by the team rather than mechanically
enforced. PR #2 is the mechanically enforced case.

## 4. PR that turned green

The same pull request, <https://github.com/DoubleHo05/LeetCpp/pull/2>:

- Commit [`9047e6e`](https://github.com/DoubleHo05/LeetCpp/commit/9047e6ec109d01b17e2247afd0616eb4c8a87c7b)
  ("Restore contracted 409 for idempotency-key-reused") ran `checks` →
  **`success`**
  ([run 42880430692](https://github.com/DoubleHo05/LeetCpp/actions/runs/42880430692)).
- Only `submit.go` changed in either commit. `submit_test.go` was never
  edited — see [`git diff`](https://github.com/DoubleHo05/LeetCpp/compare/6805acb45a398276d7ebddfdca9142b5b8156cac...9047e6ec109d01b17e2247afd0616eb4c8a87c7b)
  across the break and the fix.
- The green run and the red run above execute the same job and the same
  command: `make check` (`.github/workflows/ci.yml`, `Makefile`).

For PR #1 the green commit is
[`bb94186`](https://github.com/DoubleHo05/LeetCpp/pull/1/changes/bb94186a7064fd8306a55b804136801018de5fc5).

## 5. CI run for the tag `pa2`

The tag `pa2` points at commit **`{{TAG_SHA}}`** (the deliverables commit on
`main` — this file and the rest of the PA#2 evidence).

- The `checks` run of that exact commit:
  **`{{TAG_COMMIT_RUN}}`** (push to `main`, green).
- Pushing the tag also starts a run of the same workflow and the same
  `make check` command: **`{{TAG_PUSH_RUN}}`**.

Both URLs are printed in the submission zip at `PA2.md`. The zip is taken from
`main` at the commit that records them; the tag itself stays on the
deliverables commit so its tree is never rewritten (see
[`AI-LOG.md`](AI-LOG.md) 2026-11-11 #13). Re-running the checks on the tag is
what the brief asks graders to do, and both commits run the identical
command.

## 6. Deliberate contract break

<https://github.com/DoubleHo05/LeetCpp/pull/2> — done. The full account is in
[`AI-LOG.md`](AI-LOG.md) 2026-11-11 #10. In short:

1. `submit.go` returned `400` for `idempotency-key-reused`;
   [`docs/contract/post-labs-labid-submissions.md`](docs/contract/post-labs-labid-submissions.md)
   §5 contracts `409`.
2. Commit [`6805acb`](https://github.com/DoubleHo05/LeetCpp/commit/6805acb45a398276d7ebddfdca9142b5b8156cac),
   CI **red** — the contract test failed exactly as written:
   `submit_test.go:274: status = 400, want 409`.
3. Merge refused by the required check (see §3).
4. Commit [`9047e6e`](https://github.com/DoubleHo05/LeetCpp/commit/9047e6ec109d01b17e2247afd0616eb4c8a87c7b),
   CI **green**. No test file was edited in the break or the fix.

Both commits are on `main` (PR #2 merged with a merge commit), so the break
and the repair are visible in the history.

---

## Supporting documents

| Document | Description |
|---|---|
| [`AGENTS.md`](AGENTS.md) | Rules file: stack, commands, conventions, never; three permission tiers |
| [`AI-LOG.md`](AI-LOG.md) | Dated AI-assisted work: tools, decisions, Changed / Rejected, incidents |
| [`docs/contract/post-labs-labid-submissions.md`](docs/contract/post-labs-labid-submissions.md) | Contract for `POST /labs/{labId}/submissions` |
| [`backend/internal/api/submit_test.go`](backend/internal/api/submit_test.go) | Contract tests — 21 tests over the whole error catalog |
| [`Makefile`](Makefile) | `make check` and the individual checks |
| [`.github/workflows/ci.yml`](.github/workflows/ci.yml) | CI runs `make check` on PRs and on the `pa2` tag |
| [`README.md`](README.md) | Getting started, quality checks, secrets, spending cap, CI |
| [`SELF_ASSESSMENT_REPORT.md`](SELF_ASSESSMENT_REPORT.md) | Marks claimed per criterion, with evidence |
