# Azure Linux Web App with Docker Container and Datadog Monitoring

This example demonstrates how to deploy a Linux-based Azure Web App using a custom Docker container with Datadog monitoring enabled.

## Overview

This configuration creates:
- An Azure Resource Group
- An Azure Container Registry (ACR)
- A custom Docker image pushed to ACR
- An Azure App Service Plan (Linux, P1v2)
- An Azure Web App configured with Datadog monitoring

## Prerequisites

- Azure CLI installed and authenticated
- Docker installed and running
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

## What Gets Deployed

The example provisions a complete web application infrastructure with:
- A Linux App Service running on P1v2 tier
- Custom Docker container hosted in Azure Container Registry
- Datadog APM and monitoring with:
  - Environment set to "dev"
  - Service name "my-service"
  - Version "1.0.0"
  - Profiling enabled
- HTTPS-only access
- Custom container port configured (8080)

## Customization

You can customize the deployment by modifying:
- `datadog_env`, `datadog_service`, `datadog_version` for your environment
- `app_settings` to enable additional Datadog features
- `site_config.application_stack` for Docker image settings
- `container_config.port` for your application's listening port
- `tags` for additional resource tagging

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
| <a name="module_datadog_linux_web_app"></a> [datadog\_linux\_web\_app](#module\_datadog\_linux\_web\_app) | ../../modules/linux | n/a |

## Resources

| Name | Type |
|------|------|
| [azurerm_container_registry.example](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/container_registry) | resource |
| [azurerm_resource_group.example](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group) | resource |
| [azurerm_service_plan.example](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/service_plan) | resource |
| [terraform_data.acr_push_image](https://registry.terraform.io/providers/hashicorp/terraform/latest/docs/resources/data) | resource |

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
