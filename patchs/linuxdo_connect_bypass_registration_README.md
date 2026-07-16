# LinuxDo Connect Bypass Registration patches

Generated: 2026-07-16
Feature: linuxdo_connect_bypass_registration
Base notes:
- Local branch `main` is behind `cloud/main` by ~93 commits.
- `*_vs_cloud_main.patch` is a **three-way rebase** of the feature onto current `cloud/main`
  (not a raw `git diff cloud/main`, which would include unrelated reverse diffs).
- Docs are gitignored (`docs/*`); docs patch is separate new-file patch.

## Files

| Patch | Apply on | Content |
|-------|----------|---------|
| `linuxdo_connect_bypass_registration_vs_cloud_main.patch` | clean checkout of `cloud/main` | code + tests (25 files) |
| `linuxdo_connect_bypass_registration_docs_vs_cloud_main.patch` | same tree | analysis/proposal docs under `docs/` |
| `linuxdo_connect_bypass_registration_vs_local_HEAD.patch` | this workspace HEAD | pure uncommitted feature vs local HEAD |

## Apply (cloud/main)

```bash
git fetch cloud
git checkout -B apply-linuxdo-bypass cloud/main
git apply --index patchs/linuxdo_connect_bypass_registration_vs_cloud_main.patch
# optional docs (gitignored in this repo by default):
git apply patchs/linuxdo_connect_bypass_registration_docs_vs_cloud_main.patch
```

Dry-run:

```bash
git apply --check patchs/linuxdo_connect_bypass_registration_vs_cloud_main.patch
```

## Setting key

`linuxdo_connect_bypass_registration` (default false)