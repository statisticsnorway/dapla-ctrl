# Google AI Reconciler

Reconciles these resources for a Dapla team:

- Google Vertex AI (Gemini Enterprise Agent Platform)
- Vertex AI user IAM bindings for Dapla teams
  - `developers` group
  - Dapla Lab service account
- Cloud Billing budgets and email notification channels

These resources are currently only created in the test environment.

Reconciliation flow for a single team:

``` mermaid
flowchart LR
    A(User) -->|Enables AI for team| B(Team Features Table)
    C{Reconciler} -->|Reads table| B
    C -->|Reconciles| D[Vertex AI API]
    C -->|Reconciles| E[IAM bindings]
    C -->|Reconciles| F[Billing Budget]
```

## Required IAM permissions

The reconciler's Google service account requires these permissions on each
team's standard project:

- `resourcemanager.projects.get`
- `resourcemanager.projects.getIamPolicy`
- `resourcemanager.projects.setIamPolicy`
- `serviceusage.services.get`
- `serviceusage.services.enable`
- `serviceusage.services.disable`
- `serviceusage.operations.get`
- `monitoring.notificationChannels.list`
- `monitoring.notificationChannels.create`
- `monitoring.notificationChannels.delete`
- `billing.resourcebudgets.read`
- `billing.resourcebudgets.write`

## Tests

- `reconciler_test.go` verifies IAM safety and idempotence, notification channel convergence, budget configuration, and standard project selection.
- `fake_google_server_test.go` provides in-memory Resource Manager and Monitoring gRPC services for the tests.
