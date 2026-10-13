# Atlantis Reconciler

This reconciler manages [Atlantis](https://www.runatlantis.io/) resources for
teams that need it.

It manages:
- Terraform state buckets for each environment
- Webhook secret for [validating]((https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)) GitHub webhook events
- Kubernetes resources
  - Service account
  - Atlantis instance ([Knative Service](https://knative.dev/docs/serving/), [defaultservice.yaml.gotmpl](./defaultservice.yaml.gotmpl))
  - PVC for persistent storage 
  - ConfigMap ([Server Side Repo Config](https://www.runatlantis.io/docs/server-side-repo-config.html), [repos.yaml](./repos.yaml))
  - Secret (Webhook Secret)
- Google service account
  - With group memberships
  - Admin rights on the team's Docker [Artifact Registry](../google/artifactregistry)
  - Workload Identity binding to the corresponding Kubernetes SA
  - IAM bindings on the team's [GCP folders](../google/gcpresources)

## Requirements
In order to function correctly, the reconciler needs the follow set of permissons.

- In Google Admin (https://admin.google.com)
  - Grant the SA `Groups Admin` (must be done by a Super Admin)
- On the State Buckets projects
  - `roles/storage.admin` 
- On the Atlantis project
  - `roles/container.editor`
  - `roles/iam.serviceAccountAdmin`
- On the Artifact Registry project
  - Manage IAM bindings on all repositories
  - Should already be in place because of the AR reconciler
- On the teams' GCP folders
  - Manage IAM bindings
  - Should already be in place because of the GCP resources reconciler
