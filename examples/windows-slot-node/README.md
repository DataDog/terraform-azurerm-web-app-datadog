# Azure Windows Web App Slot with Node.js and Datadog Monitoring

This example demonstrates how to deploy a Windows-based Azure Web App **deployment slot** running Node.js 22 with Datadog monitoring enabled.

## Overview

This configuration creates:
- An Azure Resource Group
- An Azure App Service Plan (Windows, P1v2)
- A parent Azure Web App configured for Node.js 22 with Datadog monitoring
- A "staging" deployment slot on the parent app with Datadog monitoring
- Automated code deployment to the slot from local source

## Prerequisites

- Azure CLI installed and authenticated
- Node.js installed
- Terraform installed
- A Datadog account with an API key
- An Azure subscription

## Required Variables

Set the following variables in a `terraform.tfvars` file or via environment variables:

```hcl
subscription_id     = "your-azure-subscription-id"
resource_group_name = "your-resource-group-name"
name                = "your-app-name"
location            = "eastus"  # or your preferred region
datadog_api_key     = "your-datadog-api-key"
datadog_site        = "datadoghq.com"  # or your Datadog site (e.g., datadoghq.eu)
```

## Usage

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Review the planned changes:
   ```bash
   terraform plan
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

The example includes a local code deployment step that packages and deploys the sample Node.js application (index.js and package.json) to the staging slot.

## What Gets Deployed

The example provisions a complete web application infrastructure with:
- A parent Windows App Service running on P1v2 tier
- A "staging" deployment slot with:
  - Node.js 22 runtime
  - Datadog APM and monitoring with:
    - Environment set to "dev"
    - Service name "my-service"
    - Version "1.0.0"
    - Profiling enabled
  - HTTPS-only access
  - Build-on-deployment enabled for npm packages

## Customization

You can customize the deployment by modifying:
- `datadog_env`, `datadog_service`, `datadog_version` for your environment
- `app_settings` to enable additional Datadog features or application settings
- `site_config.application_stack.node_version` for a different Node.js version
- `tags` for additional resource tagging
- Replace the `terraform_data.code_deployment` resource with your preferred CI/CD deployment method

## Clean Up

To destroy all resources created by this example:
```bash
terraform destroy
```

<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_datadog_windows_web_app"></a> [datadog\_windows\_web\_app](#module\_datadog\_windows\_web\_app) | ../../modules/windows | n/a |
| <a name="module_datadog_windows_web_app_slot"></a> [datadog\_windows\_web\_app\_slot](#module\_datadog\_windows\_web\_app\_slot) | ../../modules/windows-slot | n/a |

## Resources

| Name | Type |
|------|------|
| [azurerm_resource_group.example](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group) | resource |
| [azurerm_service_plan.example](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/service_plan) | resource |
| [terraform_data.code_deployment](https://registry.terraform.io/providers/hashicorp/terraform/latest/docs/resources/data) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_datadog_api_key"></a> [datadog\_api\_key](#input\_datadog\_api\_key) | n/a | `string` | n/a | yes |
| <a name="input_datadog_site"></a> [datadog\_site](#input\_datadog\_site) | n/a | `string` | `"datadoghq.com"` | no |
| <a name="input_location"></a> [location](#input\_location) | n/a | `string` | n/a | yes |
| <a name="input_name"></a> [name](#input\_name) | n/a | `string` | n/a | yes |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | n/a | `string` | n/a | yes |
| <a name="input_subscription_id"></a> [subscription\_id](#input\_subscription\_id) | n/a | `string` | n/a | yes |

## Outputs

No outputs.
<!-- END_TF_DOCS -->
