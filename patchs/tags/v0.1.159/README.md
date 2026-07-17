# Patches for tag v0.1.159

Generated: 2026-07-17
Tag: `v0.1.159` -> commit `2a75d7d2387587d86ca3c5e5cd8ca96cf3d104c6`
Branch used for apply: `apply-patches-v0.1.159`

## Contents

| Patch | Content |
|-------|---------|
| `linuxdo_connect_bypass_registration.patch` | code + tests (25 files). LinuxDo OAuth self-registration bypass when global `registration_enabled=false`, plus UX/clarity follow-ups (always-visible admin toggle, orthogonal email-gate hints, accurate pending payload fields). |
| `linuxdo_connect_bypass_registration_docs.patch` | analysis/proposal docs under `docs/` (3 new files; usually gitignored by `docs/*`). Same content as the v0.1.155/docs-vs-cloud_main docs patch. |
| `easypay_fix.patch` | EasyPay `return_url` length guard (omit when > 100 chars) in `backend/internal/payment/provider/easypay.go`. |

## Origin

Rebased / re-applied onto `v0.1.159` from:

- `origin/dev_wtf:patchs/tags/v0.1.155/linuxdo_connect_bypass_registration.patch` (via `git apply --3way`)
- `origin/dev_wtf:patchs/tags/v0.1.155/linuxdo_connect_bypass_registration_docs.patch` (new-file docs; content unchanged across tags)
- `origin/dev_wtf:patchs/easypay_fix.patch` (clean apply)

Code/test patches were regenerated with `git diff --cached v0.1.159` so they apply **cleanly without --3way** on `v0.1.159`.

## Apply on v0.1.159

```bash
git fetch --tags
git checkout -B apply-patches-v0.1.159 v0.1.159
git apply --check patchs/tags/v0.1.159/easypay_fix.patch
git apply --check patchs/tags/v0.1.159/linuxdo_connect_bypass_registration.patch
git apply --index patchs/tags/v0.1.159/easypay_fix.patch
git apply --index patchs/tags/v0.1.159/linuxdo_connect_bypass_registration.patch
# optional docs (new files under docs/):
git apply patchs/tags/v0.1.159/linuxdo_connect_bypass_registration_docs.patch
```

## Setting key

`linuxdo_connect_bypass_registration` (default false)

Allows LinuxDo OAuth self-registration while global `registration_enabled=false`.
Does **not** bypass invitation codes, email verification, force-email-on-third-party-signup, password register, or Backend Mode.

## Notes

- Docs patch adds:
  - `docs/AUTH_REGISTRATION_LINUXDO_ANALYSIS.md`
  - `docs/LINUXDO_BYPASS_REGISTRATION_PROPOSAL.md`
  - `docs/LINUXDO_EMAIL_BYPASS_PROPOSAL.md`
- These paths are typically ignored by repo `docs/*` rules; apply only if you want the local analysis docs.
- `domain_constants.go` gained session-binding / audit-log keys between v0.1.155 and v0.1.159; the regenerated code patch includes the correct surrounding context.
