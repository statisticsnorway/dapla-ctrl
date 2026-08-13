# Google AI Reconciler

Reconciles these resources for a Dapla team:

- Google Vertex AI (Gemini Enterprise Agent Platform)
- Vertex AI user IAM bindings for Dapla teams
  - `developers` group
  - Dapla Lab service account
- Cloud Billing budgets and email notification channels

There resources are currently only created in the test environment.

## Tests

- `reconciler_test.go` verifies IAM safety and idempotence, notification channel convergence, budget configuration, and standard project selection.
- `fake_google_server_test.go` provides in-memory Resource Manager and Monitoring gRPC services for the tests.
