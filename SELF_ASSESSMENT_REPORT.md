# PA#2 — Self-assessment report

Product: **LeeTKOB** (`LeetCpp`) · CSC13114 · teams of up to 3.
Submission tag: `pa2` · Submission file name: see the zip name.

One row per rubric criterion. Every claimed mark has an evidence line pointing
at a file, a commit, a PR or a CI run. A claim with no evidence is a claim we
do not make.

| # | Criterion | Max | Claimed | Evidence |
|---|---|---|---|---|
| 1 | Rules file | 15 | **15** | `AGENTS.md` — four sections (stack, commands, conventions, never). The never list is product-specific: no secret in the sandbox, no real model call in unit tests, never edit a test / eval data / a label to make a check pass. Two lines trace to real incidents: "never modify tests just to make implementation pass" → `AI-LOG.md` 2026-10-31 #2 and #4; "never reorder the error-precedence list in code alone" → `AI-LOG.md` 2026-10-31 #5. |
| 2 | Permissions and secrets | 10 | **10** | `AGENTS.md` §3 "Agent permissions" — three tiers (allowed / ask first / never). `.env.example` lists all 10 variables with no values. `AI-LOG.md` 2026-10-31 #8 is a dated history audit: no secret has ever been committed. `README.md` §7 states the model key carries a hard **USD 15 per calendar month** cap set in the provider's billing console before the key is issued; no key exists yet and none is issued without the cap. |
| 3 | One command for every check | 10 | **10** | `Makefile` — `make check` = lint + format-check + typecheck + test, each also available alone. `.github/workflows/ci.yml` runs exactly `make check`, nothing CI-only. `README.md` §5 documents the same commands. Fresh-clone green is proven by `AI-LOG.md` 2026-11-01 #9 (CI caught a `.next/`-cache-dependent typecheck and the fix is in commit [`bb94186`](https://github.com/DoubleHo05/LeetCpp/commit/bb94186a7064fd8306a55b804136801018de5fc5)). |
| 4 | CI that blocks a merge | 20 | **20** | `main` is protected and the `checks` status is **required** — `docs/evidence/branch-protection.json` (`contexts: ["checks"]`, `strict: true`, `enforce_admins: true`, force pushes off). The gate bit for real: [PR #2](https://github.com/DoubleHo05/LeetCpp/pull/2) was refused while red with `Required status check "checks" is failing.` Red run [`42879941393`](https://github.com/DoubleHo05/LeetCpp/actions/runs/42879941393), green run [`42880430692`](https://github.com/DoubleHo05/LeetCpp/actions/runs/42880430692), fix commit [`9047e6e`](https://github.com/DoubleHo05/LeetCpp/commit/9047e6ec109d01b17e2247afd0616eb4c8a87c7b). Nothing has been pushed straight to `main`: after the initial commit, every commit on `main` arrived through PR #1 or PR #2 (merge commits `bbfdd51`, `d6620a9`). The rule was corrected as part of this submission (`AI-LOG.md` 2026-11-11 #12). |
| 5 | The contract of one LeeTKOB endpoint | 25 | **25** | `docs/contract/post-labs-labid-submissions.md` — `POST /labs/{labId}/submissions`. §2 who may call it (role `student` only, ownership rule, 404-not-403 for non-members, `studentId` never read from the client); §3 request with field rules; §4 response with field rules, 201/200 replay semantics, payload fingerprint; §5 every error as RFC 9457 problem details with the `type` URI and the status, a closed catalog of 10 types, and a 13-step error precedence; §6 idempotency (scope, retention, replay, key reuse); §7 change policy; §8 paging, correctly not applicable. A stranger could write a client from this file alone. |
| 6 | Contract tests that bite | 15 | **15** | `backend/internal/api/submit_test.go` — 21 tests, every error type in the catalog asserted on the problem `type` **and** the status, the ownership rule (`TestNonMemberGets404Not403` checks the non-member body is indistinguishable from the missing-lab body), idempotency (`TestIdempotentReplay200`, `TestIdempotencyKeyReused409`, `TestIdempotencyKeyScopedPerStudent`, `TestReplayWinsOverClosedWindow`) and error precedence. They run in the required CI (`checks`). One deliberate break is shown: commit [`6805acb`](https://github.com/DoubleHo05/LeetCpp/commit/6805acb45a398276d7ebddfdca9142b5b8156cac) violated the contract and CI went red (`submit_test.go:274: status = 400, want 409`, run [`42879941393`](https://github.com/DoubleHo05/LeetCpp/actions/runs/42879941393)); commit [`9047e6e`](https://github.com/DoubleHo05/LeetCpp/commit/9047e6ec109d01b17e2247afd0616eb4c8a87c7b) fixed the handler without touching a test file. Full account: `AI-LOG.md` 2026-11-11 #10. |
| 7 | AI-LOG.md | 5 | **5** | `AI-LOG.md` — 13 dated entries. A tools table says which tool did which part (MiMoCode, Go toolchain, pnpm toolchain, GitHub Actions, GitHub REST API). Real Changed / Rejected entries at the top and in the log: **Changed** #11 (the CI workflow could never run for the `pa2` tag), **Changed** #12 (branch protection required no status check), **Changed** #2 and #9 (test helpers, `typecheck` cache), **Rejected** #13 (force-moving the `pa2` tag). |
| | **Total** | **100** | **100** | |

## What we did not manage

- **No browser screenshot of the branch-protection settings page.** Criterion 2
  of the hand-in list says "a screenshot is fine"; we used the GitHub REST
  response instead, saved verbatim as `docs/evidence/branch-protection.json`.
  It is the machine-readable truth of the same rule and it is what the API
  refuses merges against, but it is not a picture of the Settings screen.
- **The `pa2` tag cannot contain the link to its own tag-push CI run.** The
  run URL only exists after the tag is pushed. Rather than rewrite the tag
  (`AI-LOG.md` 2026-11-11 #13, rejected), the tag points at the deliverables
  commit, `PA2.md` links the `checks` run of that exact commit, and the
  tag-push run URL is recorded in the zip's copy of `PA2.md`.
- **No model API key exists yet.** The spending cap (USD 15/month) is
  documented and will be set in the provider's billing console before the key
  is issued, but there is no live key to point at. This is deliberate: PA#3 is
  where the LLM feature arrives.
- **The frontend has one test.** The contract work for PA#2 is on the
  `POST /labs/{labId}/submissions` endpoint and its suite is the Go contract
  tests; frontend behaviour is not yet under contract.
- **Only one endpoint is contracted.** The rubric asks for one. The others
  (`GET /labs/{labId}/review-queue`, `POST /labs/{labId}/publish`,
  `POST /submissions/{id}/appeals`) are not written yet.
