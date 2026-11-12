# GCP Resources reconciler

Creates and stores team folders for each configured environment and attaches the
team tag. Folder and tag reconciliation applies to both Managed and self-managed
teams.

For Managed teams only, it also creates the `ssb-bucket-read` PAM entitlement on
each team folder in `locations/global`. The entitlement grants a configurable
IAM role to the configured eligible principals.
No approval or requester justification is required. IAM conditions are not
configured yet.

## PAM configuration

Set these environment variables for the reconciler process (or in a local
`.env` file):

| Key                                          | Value                                                                                                                             |
| -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `PAM_MANAGED_TEAM_FOLDER_PRINCIPALS`         | Comma-separated IAM principal identifiers, e.g. `serviceAccount:reader@example.iam.gserviceaccount.com,group:readers@example.com` |
| `PAM_MANAGED_TEAM_FOLDER_ROLE`               | Full IAM role name, e.g. `organizations/333333333/roles/ssb.bucket.read` or `roles/storage.objectViewer`                          |
| `PAM_MANAGED_TEAM_FOLDER_MAX_GRANT_DURATION` | Go duration between `30m` and `168h`, e.g. `1h`                                                                                   |

None of these values has a default. Until all three are set, PAM reconciliation
is skipped with a warning; folder and tag reconciliation continues. Invalid
configuration is rejected at startup, and PAM API failures are returned as
reconciliation errors. Configuration is read once at startup; restart the
process after changing environment variables. Existing entitlements are updated
when the policy changes, using an explicit field mask and the current etag;
unchanged entitlements are left alone.

PAM settings are no longer registered or read through the API-backed reconciler
configuration. Principal entries are trimmed and must not be empty; their IAM
identifier formats are validated by Google's API, not restricted to service
accounts. The entitlement ID remains `ssb-bucket-read` when its role changes so
existing entitlements are updated rather than duplicated.

Enable the Privileged Access Manager API and onboard PAM on the relevant folder
hierarchy before configuring these values. The reconciler identity needs
permissions to get, create and update PAM entitlements and the folder IAM
permissions required by PAM. See Google's
[PAM entitlement setup](https://cloud.google.com/iam/docs/pam-create-entitlements).

Deletion is not implemented. Skipping a self-managed team or clearing PAM
configuration does not delete an existing entitlement.
