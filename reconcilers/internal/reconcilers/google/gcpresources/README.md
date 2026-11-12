# GCP Resources reconciler

Creates and stores team folders for each configured environment and attaches the
team tag. Folder and tag reconciliation applies to both Managed and self-managed
teams.

`GCP_TEAM_TAG_KEY_NAMESPACED_NAME` supplies the tag key and
`GCP_TEAMS_FOLDER_NUMBERS` maps environments to their parent folder IDs.

PAM entitlements are managed separately by the
[`google:pam` reconciler](../pam), which runs after this reconciler and reads the
stored team folder IDs.
