# dapla-ctrl-labradoor

Q: `labradoor`? Shouldn't it be `labrador`?

A: No. 🚪

The app will receive commands from the reconciler and handle them appropriately.

## Configuration


| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `AUTH_TOKEN` | Yes | - | Bearer token required for authenticated API routes. Requests must include `Authorization: Bearer <token>`. |
| `PARQUEDIT_DATABASE_URL` | Yes | - | Connection string for the Parquedit database. |
| `PARQUEDIT_CLOUDSQL_CLIENT_PROJECT` | No | - | Google Cloud project where `EnableForTeam` grants `roles/cloudsql.client`. |
| `PARQUEDIT_CLOUDSQL_CLIENT_USER` | No | - | User email, or `user:<email>`, granted `roles/cloudsql.client` by `EnableForTeam`. |
| `LISTEN_ADDR` | No | `:8080` | Address the HTTP server listens on. |
