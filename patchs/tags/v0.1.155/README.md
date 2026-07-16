# LinuxDo Connect Bypass Registration — patches for tag v0.1.155

Generated: 2026-07-16
Tag: `v0.1.155` → commit `41cec0db059ffb82d0efdcfcf07a24ab51fbfe97`
Feature: `linuxdo_connect_bypass_registration` (+ UX/clarity follow-ups)

## Contents

| Patch | Content |
|-------|---------|
| `linuxdo_connect_bypass_registration.patch` | code + tests (25 files). Includes base bypass feature and UX fixes: always-visible admin toggle, orthogonal email-gate hints, accurate `force_email_on_signup` / `email_verification_required` pending payload fields. |
| `linuxdo_connect_bypass_registration_docs.patch` | analysis/proposal docs under `docs/` (gitignored in this repo) |

## Follow-ups included vs earlier v0.1.155 patch

1. **Settings UI**: `linuxdo_connect_bypass_registration` remains visible when LinuxDo master switch is off (disabled/greyed), avoiding the "toggle disappeared" confusion.
2. **Copy**: clarifies bypass only affects global registration, not email verification / force-email-on-third-party-signup.
3. **Callback payload**: when email binding is required by `email_verify_enabled` alone, `force_email_on_signup` is **false** and `email_verification_required` is **true** (no longer mislabeled).

## Apply on v0.1.155

```bash
git fetch --tags
git checkout -B apply-linuxdo-bypass v0.1.155
git apply --check patchs/tags/v0.1.155/linuxdo_connect_bypass_registration.patch
git apply --index patchs/tags/v0.1.155/linuxdo_connect_bypass_registration.patch
# optional docs:
git apply patchs/tags/v0.1.155/linuxdo_connect_bypass_registration_docs.patch
```

## Setting key

`linuxdo_connect_bypass_registration` (default false)

Allows LinuxDo OAuth self-registration while global `registration_enabled=false`.
Does **not** bypass invitation codes, email verification, force-email-on-third-party-signup, password register, or Backend Mode.
