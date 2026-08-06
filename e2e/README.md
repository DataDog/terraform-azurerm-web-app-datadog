# Azure Web App E2E test

## Lifecycle contract

The test deploys an instrumented Linux Web App, verifies its runtime, sidecar, app settings, and resource tags, invokes the workload, and waits for traces and logs. It then confirms a second Terraform plan has no changes, destroys the resources, and verifies the app is gone.

## Prerequisites

Install Go 1.23+, Terraform 1.5+, Azure CLI, and `dd-auth`. Sign in to Azure with permission to manage a resource group, App Service plan, and Linux Web App. The account also needs Storage Blob Data Reader access to the `smddsvlsprod` account, unless `E2E_WORKLOAD_ZIP` points to a local workload package.

## Run locally

```sh
cd e2e && AZURE_SUBSCRIPTION_ID="$(az account show --query id -o tsv)" dd-auth --domain ddserverless.datadoghq.com -- go test -count=1 -v -timeout 45m ./...
```

The test defaults to `eastus2`, `datadoghq.com`, and the pinned serverless-init image in `config.go`. Use `AZURE_LOCATION`, `DD_SITE`, or `E2E_SIDECAR_IMAGE` to override them.

## CI authentication and configuration

[The E2E workflow](../.github/workflows/e2e.yaml) runs for Linux module and E2E changes in the canonical repository. Fork jobs do not receive cloud credentials. Azure uses GitHub OIDC; Datadog API and application keys come from the `terraform-azurerm-web-app-datadog-e2e` dd-sts policy.

Configure these repository variables:

- `AZURE_CLIENT_ID_E2E`
- `AZURE_TENANT_ID_E2E`
- `AZURE_SUBSCRIPTION_ID_E2E`
- `DD_SITE_E2E` (optional; defaults to `datadoghq.com`)
- `E2E_STORAGE_ACCOUNT` (optional; defaults to `smddsvlsprod`)

Missing required credentials or configuration fails during authentication or the test preflight.

## Resource hygiene

Each run uses the `one-e2e-tfwebapp-linux-<run-id>` prefix. Resources carry exact `one_e2e_run_id` and `one_e2e_created` tags so the cross-repository sweeper can remove leaks. The test always attempts destroy and verifies cleanup; workflow runs are not cancelled in progress.
