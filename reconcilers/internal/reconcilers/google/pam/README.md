# PAM reconciler

The `google:pam` reconciler creates the `ssb-bucket-read` PAM entitlement on
each team folder in `locations/global`. The entitlement grants a configurable
IAM role to the configured eligible principals.
No approval or requester justification is required. A templated IAM condition
scopes the permissions granted by the entitlement.

Only Managed teams are reconciled. This reconciler runs after
[`google:gcpresources`](../gcpresources) and reads the folder IDs stored by it;
it does not create folders or tags. Environments are taken from the keys of
`GCP_TEAMS_FOLDER_NUMBERS`. Missing stored folders are returned as reconciliation
errors.

Enable `google:pam` separately through the Console, or include it in
`RECONCILERS_TO_ENABLE` before its first registration. Keep `google:gcpresources`
enabled so the team folders are created and maintained.

## PAM configuration

**API-backed configuration is the preferred method for reconciler policies.**
Set these keys for `google:pam` through the Console or the admin-only
`configureReconciler` GraphQL mutation. The reconciler registers their
specifications with Dapla API at startup.

| Key | Value |
| --- | --- |
| `principals` | Comma-separated IAM principal identifiers, e.g. `serviceAccount:reader@example.iam.gserviceaccount.com,group:readers@example.com` |
| `role` | Full IAM role name, e.g. `organizations/333333333/roles/ssb.bucket.read` or `roles/storage.objectViewer` |
| `max_grant_duration` | Go duration between `30m` and `168h`, e.g. `1h` |
| `condition_template` | Go text template that renders an IAM CEL condition; variables are `.TeamSlug` and `.Env` |
| `team_allowlist` | Optional comma-separated list of exact, case-sensitive team slugs, e.g. `my-team,another-team`. Empty or unset allows all Managed teams. |

The four policy settings have no defaults. Until all four are set, PAM reconciliation
is skipped with a warning; folder and tag reconciliation continues. Invalid
configuration and malformed templates are returned as reconciliation errors.
Template
execution failures, empty rendered conditions and PAM API failures are returned
as reconciliation errors; an unconditional entitlement is never created as a
fallback. Configuration is fetched once per Managed-team reconciliation and used
for all its environments. Changes take effect on the next reconciliation without
restarting the process. API failures are returned as errors, never replaced with
stale configuration. Existing entitlements are updated
when the policy changes, using an explicit field mask and the current etag;
unchanged entitlements are left alone.

Use `team_allowlist` to limit PAM reconciliation to selected teams during testing.
Whitespace around entries is trimmed; non-empty lists must not contain empty
entries. Matching is exact, with no wildcard or prefix matching, and the
Managed-team requirement still applies. Changes take effect on the next
reconciliation without restarting the process. Excluded teams are skipped before
folder lookup or PAM calls.

The former `PAM_MANAGED_TEAM_FOLDER_*` environment variables are no longer read.
Re-enter their values under the corresponding API keys above; environment
values are not migrated automatically or used as a fallback.

Principal entries are trimmed and must not be empty; their IAM
identifier formats are validated by Google's API, not restricted to service
accounts. The entitlement ID remains `ssb-bucket-read` when its role changes so
existing entitlements are updated rather than duplicated.

## IAM condition

Set `condition_template` to the following literal Go template (without the
Markdown fences or additional surrounding quotes). It uses Go template
`printf "%q"` to quote interpolated strings for CEL. Rendering occurs separately
for each team and environment.

```gotemplate
{{- $bucket := `(resource.type == "storage.googleapis.com/Bucket" ? resource.name.extract("projects/_/buckets/{bucket}") : resource.name.extract("projects/_/buckets/{bucket}/"))` -}}
resource.type in ["storage.googleapis.com/Bucket", "storage.googleapis.com/Object"] &&
{{$bucket}}.endsWith({{printf "%q" (printf "-%s" .Env)}}) &&
(
  {{$bucket}} == {{printf "%q" (printf "ssb-%s-data-produkt-%s" .TeamSlug .Env)}} ||
  (
    {{$bucket}}.startsWith({{printf "%q" (printf "ssb-%s-data-delt-" .TeamSlug)}}) &&
    {{$bucket}} != {{printf "%q" (printf "ssb-%s-data-delt--%s" .TeamSlug .Env)}} &&
    {{$bucket}} != {{printf "%q" (printf "ssb-%s-data-delt-delomat--%s" .TeamSlug .Env)}}
  )
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

Deletion is not implemented. Skipping a self-managed or non-allowlisted team, or
clearing PAM configuration, does not delete an existing entitlement.
