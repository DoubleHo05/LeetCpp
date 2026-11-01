# Contract — `POST /labs/{labId}/submissions`

Endpoint: a student submits one source file for a lab.
Contract version: `1` (see [Change policy](#7-change-policy)).
This document is the source of truth. The implementation and the contract tests
in `backend/internal/api/submit_test.go` must match it. If code and document
disagree, the document wins and the code is wrong.

---

## 1. Summary

| | |
|---|---|
| Method | `POST` |
| Path | `/labs/{labId}/submissions` |
| Purpose | Create one submission containing one uploaded source file for the lab identified by `labId`, for the calling student. |
| Safe to retry? | Yes. Every request carries an `Idempotency-Key`; a retried upload must never create a second submission. |
| Success | `201 Created` with the submission representation; `200 OK` on an idempotent replay. |
| Request body | `multipart/form-data` with one file part named `file`. |
| Response body | `application/json` (success) or `application/problem+json` (errors, RFC 9457). |

## 2. Who may call it

**Authentication.** The request must carry a valid session token in
`Authorization: Bearer <token>`. No token, or a token the server does not
recognize, is `401 unauthenticated`.

**Role.** Only the `student` role may call this endpoint. The roles `ta` and
`admin` are `403 role-not-allowed`, even for labs in their own classes. (TAs and
admins act on submissions through other endpoints.)

**Ownership.** The lab belongs to a class. The caller must be a member
(enrolled student) of that class:

- Lab does not exist → `404 lab-not-found`.
- Lab exists but the caller is **not** a member of the lab's class → `404 lab-not-found`.

Both cases return the *same* response body. A non-member cannot tell whether a
lab id exists. Membership is not a secret to members; lab *existence* is a
secret from non-members.

**Whose submission.** The submission is always created for the caller. There is
no request field for a student id; `studentId` in the response is derived from
the session, never read from the client.

## 3. Request

### Path parameters

| Name | Type | Rules |
|---|---|---|
| `labId` | string | Required. UUID (8-4-4-4-12 lowercase hex). Anything else is `400 validation-failed` with a field error on `labId`. |

### Headers

| Name | Required | Rules |
|---|---|---|
| `Authorization` | yes | `Bearer <session token>`. Missing/unknown → `401 unauthenticated`. |
| `Idempotency-Key` | yes | Opaque string, 1–128 characters. Missing or empty → `400 invalid-idempotency-key`. Longer than 128 → `400 invalid-idempotency-key`. |
| `Content-Type` | yes | Must be `multipart/form-data` (with a boundary). Anything else → `415 unsupported-media-type`. |

### Body — `multipart/form-data`

| Part | Rules |
|---|---|
| `file` | Required. Exactly one file part under this name is used (if a client sends several, the first is used and the rest ignored). Missing → `400 validation-failed` with a field error on `file`. |
| `file` size | At most **262144 bytes (256 KiB)**. Larger → `413 payload-too-large`. |
| `file` filename extension | Case-insensitive; one of `.cpp`, `.cc`, `.cxx`, `.h`, `.hpp`. Anything else → `422 unsupported-file-type`. |
| other parts | Ignored. |

The whole request body must not exceed **524288 bytes (512 KiB)** including
part headers. Larger → `413 payload-too-large`. A body that is not parseable as
multipart is `400 validation-failed`.

### Example

```bash
curl -X POST "https://leetc.example/labs/7c9e6679-7425-40de-944b-e07fc1f90ae7/submissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: 6f0c1a2e-6a44-4c3f-9d0a-2f8f0c1a2e6a" \
  -F "file=@sort.cpp;type=text/x-c++src"
```

## 4. Response

### `201 Created` — success

Headers:

| Name | Value |
|---|---|
| `Location` | `/submissions/{submission.id}` |
| `Content-Type` | `application/json` |

Body fields:

| Field | Type | Rules |
|---|---|---|
| `id` | string | Opaque, unique, assigned by the server. Never chosen by the client. |
| `labId` | string | Echo of the `labId` path parameter. |
| `studentId` | string | The caller's user id from the session. Not taken from the request. |
| `filename` | string | The uploaded file's name as sent in the multipart part. |
| `sizeBytes` | integer | Length of the uploaded file in bytes, 1–262144. |
| `sha256` | string | Lowercase hex SHA-256 of the uploaded file's content. |
| `status` | string | Initial value `"queued"`. A client must not assume this stays unchanged; it advances through the review pipeline. |
| `createdAt` | string | RFC 3339 timestamp in UTC, e.g. `2026-11-04T09:30:00Z`. |

```json
{
  "id": "0d5f2a9c-1d4e-4a5b-9c3e-8f7a6b5c4d3e",
  "labId": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "studentId": "u-2026-0183",
  "filename": "sort.cpp",
  "sizeBytes": 2841,
  "sha256": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
  "status": "queued",
  "createdAt": "2026-11-04T09:30:00Z"
}
```

### `200 OK` — idempotent replay

If the same caller sends the same `Idempotency-Key` again with the **same
request payload**, the server creates nothing and returns the original
submission:

- Status `200` (not `201`), body exactly the original response body.
- Header `Idempotency-Replayed: true`.
- Header `Location: /submissions/{submission.id}` still present.

### What counts as "the same payload"

The server fingerprints the request as `SHA-256(filename || 0x00 || file
bytes)`. Same fingerprint → replay. Different fingerprint under a used key →
`409 idempotency-key-reused` (see below). The key is scoped to the pair
(student, lab): two different students may use the same key value without
colliding.

## 5. Errors (RFC 9457 problem details)

Every error response has `Content-Type: application/problem+json` and this
shape:

```json
{
  "type": "https://leetc.example/problems/payload-too-large",
  "title": "Payload Too Large",
  "status": 413,
  "detail": "file part exceeds the 256 KiB limit",
  "instance": "/labs/7c9e6679-.../submissions",
  "errors": [
    { "field": "file", "message": "must be at most 262144 bytes" }
  ]
}
```

`type`, `title`, `status` are always present. `detail` is a human-readable
sentence safe to show a client developer. `instance` is the request path.
`errors` appears only on `400 validation-failed`; it lists every offending
field.

Problem `type` URIs are identifiers; a client must match on the URI string, not
on `title` or `detail` (those may be reworded). The base
`https://leetc.example/problems/` is a namespace, not a website.

### Error catalog

| `type` (suffix) | `status` | When |
|---|---|---|
| `unauthenticated` | 401 | `Authorization` header missing, malformed, or the token is not recognized. |
| `validation-failed` | 400 | `labId` is not a valid UUID; or the multipart body cannot be parsed; or the `file` part is missing. Field errors say which. |
| `invalid-idempotency-key` | 400 | `Idempotency-Key` missing, empty, or longer than 128 characters. |
| `unsupported-media-type` | 415 | `Content-Type` is not `multipart/form-data`. |
| `payload-too-large` | 413 | The request body exceeds 512 KiB, or the `file` part exceeds 256 KiB. |
| `unsupported-file-type` | 422 | The file's extension is not one of `.cpp .cc .cxx .h .hpp` (case-insensitive). |
| `lab-not-found` | 404 | The lab does not exist, **or** the caller is not a member of the lab's class. Same body in both cases. |
| `role-not-allowed` | 403 | Caller's role is `ta` or `admin`. Only `student` may submit. |
| `submission-window-closed` | 409 | The lab is not currently open: now < `opensAt` or now > `closesAt`. `detail` says which side. |
| `idempotency-key-reused` | 409 | The key was used before by this student for this lab with a **different** payload. |

### Error precedence

When a request violates more than one rule, exactly one error is returned — the
first rule below that fails:

1. `401 unauthenticated`
2. `400 validation-failed` (bad `labId`)
3. `400 invalid-idempotency-key`
4. `415 unsupported-media-type`
5. `413 payload-too-large`
6. `400 validation-failed` (missing `file` part)
7. `422 unsupported-file-type`
8. **Idempotent replay → `200 OK`** (if the key exists and the fingerprint matches)
9. `409 idempotency-key-reused`
10. `404 lab-not-found` (missing lab or non-member)
11. `403 role-not-allowed`
12. `409 submission-window-closed`
13. create → `201 Created`

Step 8 deliberately runs *before* the business rules (10–12): a retry of a
request that already succeeded returns the original result even if the lab has
since closed or the caller's membership changed. A retry never creates.

## 6. Idempotency

- The header is **required** on this endpoint, not optional. Clients generate
  one key per logical upload (UUID recommended) and reuse that key for every
  retry of the same upload (network timeout, 5xx, user double-click).
- Scope: `(studentId, labId, Idempotency-Key)`.
- Stored: the response body and the payload fingerprint, kept for at least
  **24 hours**; the server may expire keys after that (an expired key behaves
  like a new one).
- Replay: `200` + original body + `Idempotency-Replayed: true`, nothing created.
- Key reused with a different payload: `409 idempotency-key-reused` — the
  client must generate a new key for a genuinely new upload.
- There is no other deduplication: two uploads with different keys are two
  submissions, even with identical file content.

## 7. Change policy

- Clients **must ignore** unknown fields in success and error bodies.
- Additive changes (new response fields, new optional request parts, new error
  types) ship without a version bump and without notice beyond the changelog.
- Breaking changes (removing or retyping a field, changing the status code of an
  existing error type, changing an ownership or idempotency rule) require a new
  contract version: either a new media type
  (`application/vnd.leetc.submission.v2+json`) or a versioned path, chosen
  before the change ships.
- Deprecated forms keep working for at least 30 days after the changelog
  announces them; `AI-LOG.md` records the decision and the removal date.
- The error catalog above is closed for contract version `1`: a client may
  switch on the full list. New error types will only appear alongside the
  additive-change rule and will use the same problem-details shape.

## 8. Paging

Not applicable. This endpoint creates exactly one resource per call and never
returns a list. (List endpoints of this API page; their contracts document
their own `limit`/`cursor` parameters.)
