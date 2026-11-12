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

Set these environment variables for the reconciler process (or in a local
`.env` file):

| Key                                          | Value                                                                                                                             |
| -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `PAM_MANAGED_TEAM_FOLDER_PRINCIPALS`         | Comma-separated IAM principal identifiers, e.g. `serviceAccount:reader@example.iam.gserviceaccount.com,group:readers@example.com` |
| `PAM_MANAGED_TEAM_FOLDER_ROLE`               | Full IAM role name, e.g. `organizations/333333333/roles/ssb.bucket.read` or `roles/storage.objectViewer`                          |
| `PAM_MANAGED_TEAM_FOLDER_MAX_GRANT_DURATION` | Go duration between `30m` and `168h`, e.g. `1h`                                                                                   |
| `PAM_MANAGED_TEAM_FOLDER_CONDITION_TEMPLATE` | Go text template that renders an IAM CEL condition; variables are `.TeamSlug` and `.Env` |

None of these values has a default. Until all four are set, PAM reconciliation
is skipped with a warning; folder and tag reconciliation continues. Invalid
configuration and malformed templates are rejected at startup. Template
execution failures, empty rendered conditions and PAM API failures are returned
as reconciliation errors; an unconditional entitlement is never created as a
fallback. Configuration is read once at startup; restart the
process after changing environment variables. Existing entitlements are updated
when the policy changes, using an explicit field mask and the current etag;
unchanged entitlements are left alone.

PAM settings are no longer registered or read through the API-backed reconciler
configuration. Principal entries are trimmed and must not be empty; their IAM
identifier formats are validated by Google's API, not restricted to service
accounts. The entitlement ID remains `ssb-bucket-read` when its role changes so
existing entitlements are updated rather than duplicated.

## IAM condition

The single-quoted condition example in [`.env.example`](../../../../.env.example)
can be copied into a local `.env` file or supplied as the environment variable's
literal value. It uses Go template `printf "%q"` to quote interpolated strings
for CEL. Rendering occurs separately for each team and environment.

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

Deletion is not implemented. Skipping a self-managed team or clearing PAM
configuration does not delete an existing entitlement.
