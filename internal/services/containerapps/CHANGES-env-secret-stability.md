# Environment Variable and Secret Block Ordering Stability

Fixes: https://github.com/hashicorp/terraform-provider-azurerm/issues/29743
Fixes: https://github.com/hashicorp/terraform-provider-azurerm/issues/31376

## Problem

Both `azurerm_container_app` and `azurerm_container_app_job` produced noisy
remove/re-add diffs on every `terraform plan` when environment variables or
secrets were defined. The root causes were:

1. **env blocks used TypeList** — any reordering by the API (or by the user in
   HCL) caused index-based churn.
2. **Secret optional fields had no Default** — Terraform stored `null` in state
   but the SDK returned `""` after read, causing the entire set to be replaced.
3. **Set-level Sensitive on secrets** — prevented Terraform from showing *which*
   secret changed, making diffs unreadable.
4. **API doesn't return secret values** — every refresh showed `value=""` unless
   prior state values were carried forward.

## Changes

### Schema (`helpers/container_apps.go`)

| Block  | Change | Why |
|--------|--------|-----|
| `env`  | `TypeList` → `TypeSet` with `containerEnvVarHash` | Order-independent identity |
| `env`  | Removed `MinItems: 1` | Prevented phantom empty entries |
| `env.value` | Added `DiffSuppressFunc` | Irrelevant when `secret_name` is set |
| `env.secret_name` | Added `DiffSuppressFunc` | Irrelevant when `value` is set; also suppresses nil-vs-"" |
| `secret` | Added `containerSecretHash` (name-only) | Stable identity regardless of value changes |
| `secret` | Removed set-level `Sensitive` | Only `secret.value` remains sensitive |
| `secret.identity` | Added `Default:""` + case-insensitive `DiffSuppressFunc` | Prevents null-vs-"" and casing diffs |
| `secret.key_vault_secret_id` | Added `Default:""` | Prevents null-vs-"" diffs |
| `secret.value` | Added `Default:""` | Prevents null-vs-"" diffs |

### Flatten/Preserve (`helpers/container_apps.go`, `helpers/container_app_job.go`)

- `FlattenContainerAppSecrets` / `FlattenContainerAppJobSecrets`: Uses
  `strings.TrimSpace` on KeyVaultURL and Identity to normalize API blanks.
- `PreserveContainerAppSecretValues`: Carries forward plain-text secret values
  from prior state into current state (API never returns them).

### Hash functions (`helpers/container_apps.go`)

- `containerEnvVarHash`: Hashes `name + type_discriminator + value_or_secret`.
  The type discriminator (`"value"` or `"secret"`) prevents collisions when a
  plain value equals a secret name.
- `containerSecretHash`: Hashes name only (value not available from API).

### State Migration (`migration/container_app_v0_to_v1.go`, `migration/container_app_job_v0_to_v1.go`)

- Provides static v0 schema snapshots (`envSchemaV0`, `secretSchemaV0`) so
  Terraform can decode existing state.
- `UpgradeFunc` normalizes `nil` → `""` for all fields that gained `Default:""`.
- Migration files are point-in-time snapshots and impose no maintenance burden.

### Resource files

- `container_app_resource.go` / `container_app_job_resource.go`: Added
  `StateUpgraders` entry and call to `PreserveContainerAppSecretValues` during
  read to merge prior state secret values.

## Testing

- Unit tests for hash functions (including collision resistance).
- Unit tests for `FlattenContainerAppSecrets` and `PreserveContainerAppSecretValues`.
- Acceptance tests verifying plan stability after env/secret reordering.

## Upgrade path

Existing users will see a one-time state migration on `terraform plan` after
upgrading the provider. Running `terraform apply -refresh-only -auto-approve`
persists the migrated state. No manual intervention required.
