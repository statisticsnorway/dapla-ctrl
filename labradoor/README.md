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
| `PARQUEDIT_STORAGE_PATH` | No | `gs://ssb-{team}-data-produkt-test/.parquedit_data/` | DuckLake GCS data path. Must contain `{team}` and end with `/`. Set the production path explicitly in the production deployment. |
| `WITH_FAKE_CLOUD_RESOURCE_MANAGER` | No | `false` | Use the fake Cloud Resource Manager implementation, which just return without any side effects. |
| `WITH_FAKE_SQL_ADMIN` | No | `false` | Use the fake SQL Admin implementation, which will work against the local DB. |
| `FAKE_SQL_ADMIN_DB_CONN_STRING` | No | - | Database connection string used when fake SQL Admin is enabled. |

## Production and test

Run separate Labradoor deployments, each with its own `AUTH_TOKEN`,
`PARQUEDIT_DATABASE_URL`, Cloud SQL project/instance and developer service-account
suffix. The reconciler sends the same team enable/disable setting to both; each
deployment modifies only its configured database.

Set the production path explicitly on the production deployment
(`https://labradoor.lab.dapla.ssb.no`):

```dotenv
PARQUEDIT_STORAGE_PATH=gs://ssb-{team}-data-produkt-prod/.parquedit_data/
```

The test deployment (`https://labradoor.lab.dapla-test.ssb.no`) uses this default:

```dotenv
PARQUEDIT_STORAGE_PATH=gs://ssb-{team}-data-produkt-test/.parquedit_data/
```
