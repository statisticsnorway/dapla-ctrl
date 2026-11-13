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
| `PARQUEDIT_DATA_PATH_TEMPLATE` | No | `gs://ssb-{team}-data-produkt-prod/.parquedit_data/` | DuckLake GCS data path. Must contain `{team}` and end with `/`. Set an explicit isolated path in test deployments.gs://ssb-{team}-data-produkt-test/.parquedit_data/ |
| `WITH_FAKE_CLOUD_RESOURCE_MANAGER` | No | `false` | Use the fake Cloud Resource Manager implementation, which just return without any side effects. |
| `WITH_FAKE_SQL_ADMIN` | No | `false` | Use the fake SQL Admin implementation, which will work against the local DB. |
| `FAKE_SQL_ADMIN_DB_CONN_STRING` | No | - | Database connection string used when fake SQL Admin is enabled. |

## Production and test deployments

Use the same application and routes for both deployments. Each deployment manages
only the database and Cloud SQL project/instance configured through its environment.
Use separate database connections and authentication tokens for production and test.

Production retains its existing storage path by default:

```dotenv
PARQUEDIT_DATA_PATH_TEMPLATE=gs://ssb-{team}-data-produkt-prod/.parquedit_data/
```

In the test deployment, explicitly configure the per-team test bucket:

```dotenv
PARQUEDIT_DATA_PATH_TEMPLATE=gs://ssb-{team}-data-produkt-test/.parquedit_data/
```

`{team}` is replaced with the team slug. Invalid templates fail application startup.
Moving Labradoor to a test project does not change the default production GCS path;
the explicit test override is required to isolate storage. No GCS bucket is created
or its permissions changed by this setting.

Use separate team developer service accounts for production and test. Set a different
`PARQUEDIT_CLOUDSQL_USER_SUFFIX` in each deployment: the account email is the team
slug followed by that suffix. The configured suffix is used consistently for IAM
grants, Cloud SQL users, SQL privileges and cleanup; Labradoor does not create the
service accounts themselves.

Different suffixes select different accounts, but do not by themselves prevent
cross-environment access. Keep each account's IAM permissions restricted to its
intended environment, and scope Labradoor's runtime identity to its deployment.

The configured data path is stored when enabling a team; changing the environment
variable does not automatically migrate existing schemas or move GCS data. Disabling
a team drops its configured database schema, but does not delete GCS objects.
