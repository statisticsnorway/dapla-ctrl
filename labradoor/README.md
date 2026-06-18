# dapla-ctrl-labradoor

Q: `labradoor`? Shouldn't it be `labrador`?

A: No. 🚪

The app will receive commands from the reconciler and handle them appropriately.

## Configuration

Environment variables can be provided via a local `.env` file for development.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `AUTH_TOKEN` | Yes | - | Bearer token required for authenticated API routes. Requests must include `Authorization: Bearer <token>`. |
| `LISTEN_ADDR` | No | `:8080` | Address the HTTP server listens on. |
| `PARQUEDIT_DATABASE_URL` | Yes | - | Connection string for the Parquedit database. |
| `PARQUEDIT_CLOUDSQL_PROJECT` | No | - | Google Cloud project containing the Cloud SQL instance. |
| `PARQUEDIT_CLOUDSQL_INSTANCE` | No | - | Cloud SQL instance where the app adds/removes IAM service-account users and creates/drops team schemas. |
| `PARQUEDIT_CLOUDSQL_USER_SUFFIX` | No | - | Suffix used when constructing Cloud SQL IAM service-account users, for example `-developers@my-project-1d.iam.gserviceaccount.com`. Should be empty when running on local machine. |
| `WITH_FAKE_CLOUD_RESOURCE_MANAGER` | No | `false` | Use the fake Cloud Resource Manager implementation, which just return without any side effects. |
| `WITH_FAKE_SQL_ADMIN` | No | `false` | Use the fake SQL Admin implementation, which will work against the local DB. |
| `FAKE_SQL_ADMIN_DB_CONN_STRING` | No | - | Database connection string used when fake SQL Admin is enabled. |
