# E2E Conformance Suite

End-to-end integration tests for the Datadog Azure Web App Terraform module,
conforming to the contract in
[`serverless-ci/e2e/spec.md`](https://github.com/DataDog/serverless-ci/blob/main/e2e/spec.md).

The suite provisions a **real** ephemeral Azure Linux Web App, instruments it
with the module, proves telemetry flows to Datadog, asserts idempotency, then
tears everything down. It is the Go + Terratest analogue of the `datadog-ci`
Jest reference impl (`e2e/cloud-run.test.ts` + `e2e/helpers/*`).

## Lifecycle

This module is a **wrapper**: `terraform apply` is the instrumentation
mechanism and `terraform destroy` is removal. The spec lifecycle maps as:

| Spec step                          | Here                                                            |
| ---------------------------------- | --------------------------------------------------------------- |
| provision uninstrumented workload  | host (RG + plan) created by the same apply; workload code deployed after |
| **APPLY** → verify CONFIG          | `terraform apply` the fixture → assert sidecar + DD_* settings + tags    |
| trigger → verify TELEMETRY         | deploy prebuilt app, HTTP GET, poll Datadog for traces + logs            |
| **APPLY again** → idempotent       | `terraform plan -detailed-exitcode` must report no changes               |
| **REMOVE** → verify CLEAN          | `terraform destroy` → assert the web app no longer exists                |
| teardown (always)                  | `defer terraform.Destroy` runs on success and failure alike              |

## Conformance assertions

- **Config** (`internal/verify`): `datadog-sidecar` sitecontainer pinned to the
  expected `serverless-init` image on port `8126`; `DD_API_KEY` wired;
  `DD_SITE` / `DD_SERVICE` / `DD_ENV` / `DD_VERSION` /
  `WEBSITES_ENABLE_APP_SERVICE_STORAGE` set to the expected **values**; DD
  resource tags (`service` / `env` / `version` / `dd_sls_terraform_module`).
- **Telemetry** (`internal/telemetry`): polls APM spans and logs (20s × 30,
  ~10 min), filtered by the run-unique `service` + `env` — baked into the query,
  so a hit proves that identity on ingested telemetry (**identity, not
  existence**). The run-unique service doubles as the run-id marker. The workload
  logs to stdout, which Linux App Service writes to a per-instance file on the
  `/home` volume shared with the sidecar; the fixture sets
  `DD_AAS_INSTANCE_LOGGING_ENABLED` so `serverless-init` tails it, and the suite
  drives continuous traffic during the poll so the end-tailer always has fresh
  lines. Logs are therefore required alongside traces.
- **Clean end-state**: after destroy the web app must be gone (explicit absence).

### Scope

- **Linux only.** Windows App Service has no `serverless-init` log support, so it
  is not tested here. A Windows suite is a separate follow-up; it would reuse
  `Options.ExpectLogs: false` (traces only).

### Known gaps (verified against a live run, 2026-06-16)

- **`version` on spans**: `DD_VERSION` reaches the app and the `version`
  resource tag is applied (both asserted at the config layer), but the value was
  not observed on ingested spans, so span `version` identity is not asserted
  (config-layer version identity stands in).

## Resource hygiene

Every resource is created with name prefix `one-e2e-tfwebapp-linux-<runid>` and a
`one_e2e_created:<unix-ts>` freshness tag plus a `one_e2e_run_id` marker, set
atomically at creation (`internal/naming`). This is the identity the cross-repo
sweeper keys on. Artifact versions (sidecar image) are pinned so a failure
blames the module, not upstream.

## Workload app

The workload is the prebuilt **prod** Node.js sidecar-flavor package
(`node-sidecar.zip`) published by `serverless-init-self-monitoring` to the
`code` container of the `smddsvlsprod` storage account. It is **pulled, never
rebuilt** (`az storage blob download` + `az webapp deploy`).

## Running locally

Prerequisites:

- **Terraform** ≥ 1.5, **Go** ≥ 1.23, **Azure CLI** (`az`).
- **Azure auth**: `az login` to a subscription that can create resource groups,
  app service plans, Linux Web Apps, and **read** the `smddsvlsprod` storage
  account's `code` container (Storage Blob Data Reader).
- **Datadog**: `dd-auth` mints short-lived keys for the org -- no pasted keys. It
  injects `$DD_API_KEY` (also the key wired into the workload) and `$DD_APP_KEY`
  into the wrapped command, which the suite maps to `DATADOG_API_KEY` /
  `DATADOG_APP_KEY` for telemetry queries.

```sh
cd e2e

export AZURE_SUBSCRIPTION_ID="$(az account show --query id -o tsv)"
export DD_SITE=datadoghq.com
# optional overrides
export E2E_STORAGE_ACCOUNT=smddsvlsprod
export E2E_SIDECAR_IMAGE="index.docker.io/datadog/serverless-init:<pinned-tag>"

# Datadog auth: dd-auth mints short-lived keys for the org -- no pasted keys.
# It injects $DD_API_KEY (also the workload key) and $DD_APP_KEY into the wrapped command.
dd-auth --domain app.datadoghq.com -- bash -c '
  export DATADOG_API_KEY="$DD_API_KEY" DATADOG_APP_KEY="$DD_APP_KEY"
  GO111MODULE=on go test -v -timeout 45m ./...
'
```

**Without storage RBAC.** If your principal can't read the `smddsvlsprod`
artifact account, reconstruct the workload package from the self-monitoring
source (identical to what the publish pipeline zips) and point `E2E_WORKLOAD_ZIP`
at it; CI keeps pulling the prebuilt artifact:

```sh
SRC=.../serverless-init-self-monitoring/apps/code/sidecar/node
( cd "$SRC" && npm install --omit=dev && zip -rq /tmp/node-sidecar.zip . )
export E2E_WORKLOAD_ZIP=/tmp/node-sidecar.zip
```

> This repository lives under `$GOPATH`, so `GO111MODULE=on` is required for
> module-mode resolution.

Set `SKIP_AAS_TESTS=true` to skip (the test records a green skip).

## CI

`.github/workflows/e2e.yaml` runs the suite on PRs/pushes that touch
`modules/**` or `e2e/**`, gated by a `dorny/paths-filter` job that drives
`SKIP_AAS_TESTS`. Auth is GitHub → Azure **OIDC federation** (`azure/login` +
`ARM_USE_OIDC`); no long-lived credentials.

Until the OIDC federation and e2e vars/secrets below are configured, the job
green-skips the cloud lifecycle (it still compiles the suite and runs the unit
tests). It runs the real lifecycle automatically once `AZURE_CLIENT_ID_E2E` is
set. Required repo config:

| Kind   | Name                                                    | Purpose                       |
| ------ | ------------------------------------------------------- | ----------------------------- |
| var    | `AZURE_CLIENT_ID_E2E` / `AZURE_TENANT_ID_E2E` / `AZURE_SUBSCRIPTION_ID_E2E` | OIDC federation target |
| var    | `DD_SITE_E2E`                                           | Datadog site                  |
| var    | `E2E_STORAGE_ACCOUNT` / `E2E_SIDECAR_IMAGE`             | workload artifact + pin       |
| secret | `DD_API_KEY_E2E`                                        | key wired into the workload   |
| secret | `DATADOG_API_KEY_E2E` / `DATADOG_APP_KEY_E2E`           | telemetry query credentials   |

## Layout

```
e2e/
  webapp_e2e_test.go        # lifecycle orchestration (mirrors cloud-run.test.ts)
  fixtures/linux-node/      # terraform fixture: RG + plan + module + workload host
  internal/
    naming/                 # one-e2e naming + freshness-tag convention
    exec/                   # bounded-retry exec helper (mirrors exec.ts)
    azure/                  # az CLI wrappers (app settings, sitecontainers, deploy)
    verify/                 # config verifier (mirrors aas-verifier.ts)
    telemetry/              # spans + logs poller (mirrors *-telemetry-checker.ts)
```

Helpers under `internal/` are runner-agnostic (no test framework imports) so the
verification logic stays reusable, matching the spec's guidance.
