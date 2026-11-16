# Metamapper reconciler

The implementation lives in `metadata/metamapper`. Its API identity is
`metadata:metamapper:pam`.

The `metadata:metamapper:pam` reconciler creates a PAM entitlement on
each team folder in `locations/global`. The entitlement grants a configurable
IAM role to the configured eligible principals.
No approval or requester justification is required. A templated IAM condition
scopes the permissions granted by the entitlement.

Only Managed teams are reconciled. This reconciler runs after
[`google:gcpresources`](../../google/gcpresources) and reads the folder IDs stored by it;
it does not create folders or tags. Environments are taken from the keys of
`GCP_TEAMS_FOLDER_NUMBERS`. Missing stored folders are returned as reconciliation
errors.

Enable `metadata:metamapper:pam` separately through the Console, or include it in
`RECONCILERS_TO_ENABLE` before its first registration. Keep `google:gcpresources`
enabled so the team folders are created and maintained.

## PAM configuration

**API-backed configuration is the preferred method for reconciler policies.**
Set these keys for `metadata:metamapper:pam` through the Console or the admin-only
`configureReconciler` GraphQL mutation. The reconciler registers their
specifications with Dapla API at startup.

Each environment has independent configuration, with no shared defaults or
fallback between TEST and PROD:

| TEST key | PROD key | Value |
| --- | --- | --- |
| `test_principals` | `prod_principals` | Comma-separated IAM principal identifiers, e.g. `user:reader@example.com,group:readers@example.com` |
| `test_role` | `prod_role` | Full IAM role name, e.g. `roles/storage.objectViewer` |
| `test_max_grant_duration` | `prod_max_grant_duration` | Go duration between `30m` and `168h`, e.g. `1h` |
| `test_condition_template` | `prod_condition_template` | Go text template rendering an IAM CEL condition using `.TeamSlug` only |
| `test_team_allowlist` | `prod_team_allowlist` | Optional comma-separated list of exact, case-sensitive team slugs. Empty or unset allows all Managed teams in that environment. |

Keys are registered for environments in `GCP_TEAMS_FOLDER_NUMBERS`, using the
lowercase environment name as their prefix. Folder lookups retain the original
environment name, e.g. `TEST` or `PROD`.

The four policy settings have no defaults. Until all four are set for an
environment, PAM reconciliation for that environment is skipped with a warning;
the other environment and folder/tag reconciliation continue. Invalid
configuration and malformed templates are returned as reconciliation errors.
Template
execution failures, empty rendered conditions and PAM API failures are returned
as reconciliation errors; an unconditional entitlement is never created as a
fallback. Configuration is fetched once per Managed-team reconciliation, with
each environment using its own policy from that snapshot. Changes take effect on the next reconciliation without
restarting the process. API failures are returned as errors, never replaced with
stale configuration. Existing entitlements are updated
when the policy changes, using an explicit field mask and the current etag;
unchanged entitlements are left alone.

Use `test_team_allowlist` and `prod_team_allowlist` to limit PAM reconciliation
independently during testing. The Managed-team requirement still applies.
For each team allowlist, whitespace around entries is trimmed and non-empty lists
must not contain empty entries. Matching is exact and case-sensitive, with no
wildcard or prefix matching. Changes take effect on the next reconciliation
without restarting the process. Excluded teams and unconfigured environments are skipped
before folder lookup or PAM calls.

The former unprefixed API policy keys and `environment_allowlist` are removed.
Re-enter policy values separately under the TEST and PROD keys; values are not
copied automatically. To disable one environment, clear any of its four required
policy settings. Existing entitlements are not deleted.

The former `PAM_MANAGED_TEAM_FOLDER_*` environment variables are no longer read.
Re-enter their values under the corresponding API keys above; environment
values are not migrated automatically or used as a fallback.

Principal entries are trimmed and must not be empty; their IAM
identifier formats are validated by Google's API, not restricted to service
accounts. The entitlement ID remains `ssb-bucket-read` when its role changes so
existing entitlements are updated rather than duplicated.

## IAM condition

Set `test_condition_template` and/or `prod_condition_template` to the following
literal Go template (without the
Markdown fences or additional surrounding quotes). It uses Go template
`printf "%q"` to quote interpolated strings for CEL. Rendering occurs separately
for each team and environment.

```gotemplate
resource.type in ["storage.googleapis.com/Bucket", "storage.googleapis.com/Object"] &&
(
  resource.name.startsWith({{printf "%q" (printf "projects/_/buckets/ssb-%s-data-produkt" .TeamSlug)}}) || resource.name.startsWith({{printf "%q" (printf "projects/_/buckets/ssb-%s-data-delt" .TeamSlug)}})
)
```

The example allows the following bucket names for that team and environment:

- `ssb-<team>-data-produkt-<env>`
- `ssb-<team>-data-delt-<kortnavn>-<env>`
- `ssb-<team>-data-delt-delomat-<kortnavn>-<env>`

The `data-delt-` prefix covers both shared-data formats, with non-empty short
names. `ssb-<team>-data-kilde-<env>` and other bucket families do not match the
allowlist. Other teams and environments do not match either.

The condition applies only to Cloud Storage Buckets and Objects. Bucket-level
permissions in the configured role, including object listing, are allowed only
for matching buckets. It extracts the bucket name before checking prefixes and
the environment suffix, so object names cannot spoof a permitted bucket.
The entitlement remains attached to the team's environment folder, limiting
its inherited permissions to resources within that folder. Bucket names, not
owning project names, are checked by the condition.

Custom condition expressions are validated by Google's PAM API. The example
uses supported IAM `extract`, `startsWith` and `endsWith` functions rather than
regular expressions or wildcards. See the
[IAM resource-name condition reference](https://cloud.google.com/iam/docs/conditions-attribute-reference#resource-name).
This condition limits this entitlement's grants; it is not a deny policy and
does not revoke access granted independently by other IAM bindings.

Enable the Privileged Access Manager API and onboard PAM on the relevant folder
hierarchy before configuring these values. The reconciler identity needs
permissions to get, create and update PAM entitlements and the folder IAM
permissions required by PAM. See Google's
[PAM entitlement setup](https://cloud.google.com/iam/docs/pam-create-entitlements).

Deletion is not implemented. Skipping a self-managed or non-allowlisted team,
excluding an environment, or clearing PAM configuration does not delete an
existing entitlement.
