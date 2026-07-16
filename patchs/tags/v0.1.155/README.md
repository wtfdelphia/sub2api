# LinuxDo Connect Bypass Registration — patches for tag v0.1.155

Generated: 2026-07-16
Tag: `v0.1.155` → commit `41cec0db059ffb82d0efdcfcf07a24ab51fbfe97`
Feature: `linuxdo_connect_bypass_registration`

## Why this is a clean tag patch

Local `HEAD` and tag `v0.1.155` share identical **committed** content for all
feature-touched files. The working-tree change is therefore a pure feature delta
vs the tag (no three-way rebase needed, unlike `cloud/main` which diverged).

## Files

| Patch | Content |
|-------|---------|
| `linuxdo_connect_bypass_registration.patch` | code + tests (25 files, includes new settings handler test) |
| `linuxdo_connect_bypass_registration_docs.patch` | analysis/proposal docs under `docs/` (gitignored in this repo) |

## Apply on v0.1.155

```bash
git fetch --tags cloud
git checkout -B apply-linuxdo-bypass v0.1.155
git apply --check patchs/tags/v0.1.155/linuxdo_connect_bypass_registration.patch
git apply --index patchs/tags/v0.1.155/linuxdo_connect_bypass_registration.patch
# optional docs:
git apply patchs/tags/v0.1.155/linuxdo_connect_bypass_registration_docs.patch
```

## Setting key

`linuxdo_connect_bypass_registration` (default false)

Allows LinuxDo OAuth self-registration while global `registration_enabled=false`.
Does **not** bypass invitation codes, email verification, password register, or Backend Mode.