# Datadog Azure Windows Web App Terraform Module

> **Technical Preview**: This module is in technical preview. While it is functional, we recommend validating it in your environment before widespread use.
> If you encounter any issues, please open a GitHub issue to let us know.

Use [this Terraform module](https://registry.terraform.io/modules/DataDog/web-app-datadog/azurerm/latest/submodules/windows) to deploy an Azure Windows Web App with integrated Datadog monitoring.

This module wraps the [azurerm_windows_web_app](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/windows_web_app) resource and configures the Datadog extension using [azapi_resource](https://registry.terraform.io/providers/Azure/azapi/latest/docs/resources/resource). It provides a simple interface to enable Datadog monitoring, Unified Service Tagging, and other best practices for observability on Azure App Service (Windows).

## Usage

```hcl
module "windows_web_app_datadog" {
  source = "DataDog/web-app-datadog/azurerm//modules/windows"

  name                = "example-app"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  service_plan_id     = azurerm_service_plan.example.id

  datadog_api_key     = var.datadog_api_key
  datadog_site        = "datadoghq.com" # or your Datadog site
  datadog_env         = "prod"
  datadog_service     = "my-dotnet-app"
  datadog_version     = "1.0.0"

  site_config = {
    application_stack = {
      dotnet_version = "v6.0"
    }
    always_on = true
  }

  app_settings = {
    WEBSITE_RUN_FROM_PACKAGE = 1
  }

  zip_deploy_file = "./src/code.zip"
}
```

## Configuration

### Azure Windows Web App

This module exposes all supported arguments available in the [azurerm_windows_web_app](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/windows_web_app) resource. Configuration blocks are provided as nested objects or maps, as shown in the example above. See the [Inputs](#inputs) section for all supported variables.

### Datadog Integration

- **API Key**: The `datadog_api_key` variable is required. You can generate an API key in the [Datadog API Keys page](https://app.datadoghq.com/organization-settings/api-keys).
- **Site**: The `datadog_site` variable defaults to `datadoghq.com`. Set this to your Datadog region if needed (e.g., `datadoghq.eu`).
- **Unified Service Tagging**: Use `datadog_env`, `datadog_service`, and `datadog_version` to enable [Unified Service Tagging](https://docs.datadoghq.com/getting_started/tagging/unified_service_tagging/).
- **App Settings**: If you use `zip_deploy_file`, set `WEBSITE_RUN_FROM_PACKAGE=1` in `app_settings`.

The module will automatically configure the Datadog extension for your web app, enabling metrics, traces, and logs collection.

### Example: Minimal Configuration

```hcl
module "windows_web_app_datadog" {
  source = "DataDog/web-app-datadog/azurerm//modules/windows"

  name                = "my-app"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  service_plan_id     = azurerm_service_plan.sp.id
  datadog_api_key     = var.datadog_api_key
  site_config = {
    application_stack = {
      dotnet_version = "v6.0"
    }
  }
}
```

### Example: Deploying a ZIP Package

```hcl
module "windows_web_app_datadog" {
  source = "DataDog/web-app-datadog/azurerm//modules/windows"

  name                = "my-app"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  service_plan_id     = azurerm_service_plan.sp.id
  datadog_api_key     = var.datadog_api_key
  zip_deploy_file     = "./src/code.zip"
  app_settings = {
    WEBSITE_RUN_FROM_PACKAGE = 1
  }
  site_config = {
    application_stack = {
      dotnet_version = "v6.0"
    }
  }
}
```

### Datadog Documentation

- [Azure App Service for Windows](https://docs.datadoghq.com/serverless/azure_app_services/azure_app_services_windows)
- [Unified Service Tagging](https://docs.datadoghq.com/getting_started/tagging/unified_service_tagging/)

---

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.0 |
| <a name="requirement_azapi"></a> [azapi](#requirement\_azapi) | >= 2.5.0 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | >= 4.57.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [azapi_resource.datadog_extension](https://registry.terraform.io/providers/Azure/azapi/latest/docs/resources/resource) | resource |
| [azurerm_windows_web_app.this](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/windows_web_app) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_app_settings"></a> [app\_settings](#input\_app\_settings) | A map of key-value pairs of App Settings. | `map(string)` | `null` | no |
| <a name="input_auth_settings"></a> [auth\_settings](#input\_auth\_settings) | n/a | <pre>object({<br/>    additional_login_parameters   = optional(map(string)),<br/>    enabled                       = bool,<br/>    issuer                        = optional(string),<br/>    token_refresh_extension_hours = optional(number),<br/>    token_store_enabled           = optional(bool),<br/>    active_directory = optional(object({<br/>      allowed_audiences          = optional(list(string)),<br/>      client_id                  = string,<br/>      client_secret              = optional(string),<br/>      client_secret_setting_name = optional(string)<br/>    })),<br/>    facebook = optional(object({<br/>      app_id                  = string,<br/>      app_secret              = optional(string),<br/>      app_secret_setting_name = optional(string),<br/>      oauth_scopes            = optional(list(string))<br/>    })),<br/>    github = optional(object({<br/>      client_id                  = string,<br/>      client_secret              = optional(string),<br/>      client_secret_setting_name = optional(string),<br/>      oauth_scopes               = optional(list(string))<br/>    })),<br/>    google = optional(object({<br/>      client_id                  = string,<br/>      client_secret              = optional(string),<br/>      client_secret_setting_name = optional(string),<br/>      oauth_scopes               = optional(list(string))<br/>    })),<br/>    microsoft = optional(object({<br/>      client_id                  = string,<br/>      client_secret              = optional(string),<br/>      client_secret_setting_name = optional(string),<br/>      oauth_scopes               = optional(list(string))<br/>    })),<br/>    twitter = optional(object({<br/>      consumer_key                 = string,<br/>      consumer_secret              = optional(string),<br/>      consumer_secret_setting_name = optional(string)<br/>    }))<br/>  })</pre> | `null` | no |
| <a name="input_auth_settings_v2"></a> [auth\_settings\_v2](#input\_auth\_settings\_v2) | n/a | <pre>object({<br/>    auth_enabled                            = optional(bool),<br/>    config_file_path                        = optional(string),<br/>    default_provider                        = optional(string),<br/>    excluded_paths                          = optional(list(string)),<br/>    forward_proxy_convention                = optional(string),<br/>    forward_proxy_custom_host_header_name   = optional(string),<br/>    forward_proxy_custom_scheme_header_name = optional(string),<br/>    http_route_api_prefix                   = optional(string),<br/>    require_authentication                  = optional(bool),<br/>    require_https                           = optional(bool),<br/>    runtime_version                         = optional(string),<br/>    unauthenticated_action                  = optional(string),<br/>    active_directory_v2 = optional(object({<br/>      allowed_applications                 = optional(list(string)),<br/>      allowed_audiences                    = optional(list(string)),<br/>      allowed_groups                       = optional(list(string)),<br/>      allowed_identities                   = optional(list(string)),<br/>      client_id                            = string,<br/>      client_secret_certificate_thumbprint = optional(string),<br/>      client_secret_setting_name           = optional(string),<br/>      jwt_allowed_client_applications      = optional(list(string)),<br/>      jwt_allowed_groups                   = optional(list(string)),<br/>      login_parameters                     = optional(map(string)),<br/>      tenant_auth_endpoint                 = string,<br/>      www_authentication_disabled          = optional(bool)<br/>    })),<br/>    apple_v2 = optional(object({<br/>      client_id                  = string,<br/>      client_secret_setting_name = string<br/>    })),<br/>    azure_static_web_app_v2 = optional(object({<br/>      client_id = string<br/>    })),<br/>    custom_oidc_v2 = optional(list(object({<br/>      client_id                     = string,<br/>      name                          = string,<br/>      name_claim_type               = optional(string),<br/>      openid_configuration_endpoint = string,<br/>      scopes                        = optional(list(string))<br/>    }))),<br/>    facebook_v2 = optional(object({<br/>      app_id                  = string,<br/>      app_secret_setting_name = string,<br/>      login_scopes            = optional(list(string))<br/>    })),<br/>    github_v2 = optional(object({<br/>      client_id                  = string,<br/>      client_secret_setting_name = string,<br/>      login_scopes               = optional(list(string))<br/>    })),<br/>    google_v2 = optional(object({<br/>      allowed_audiences          = optional(list(string)),<br/>      client_id                  = string,<br/>      client_secret_setting_name = string,<br/>      login_scopes               = optional(list(string))<br/>    })),<br/>    login = object({<br/>      allowed_external_redirect_urls    = optional(list(string)),<br/>      cookie_expiration_convention      = optional(string),<br/>      cookie_expiration_time            = optional(string),<br/>      logout_endpoint                   = optional(string),<br/>      nonce_expiration_time             = optional(string),<br/>      preserve_url_fragments_for_logins = optional(bool),<br/>      token_refresh_extension_time      = optional(number),<br/>      token_store_enabled               = optional(bool),<br/>      token_store_path                  = optional(string),<br/>      token_store_sas_setting_name      = optional(string),<br/>      validate_nonce                    = optional(bool)<br/>    }),<br/>    microsoft_v2 = optional(object({<br/>      allowed_audiences          = optional(list(string)),<br/>      client_id                  = string,<br/>      client_secret_setting_name = string,<br/>      login_scopes               = optional(list(string))<br/>    })),<br/>    twitter_v2 = optional(object({<br/>      consumer_key                 = string,<br/>      consumer_secret_setting_name = string<br/>    }))<br/>  })</pre> | `null` | no |
| <a name="input_backup"></a> [backup](#input\_backup) | n/a | <pre>object({<br/>    enabled             = optional(bool),<br/>    name                = string,<br/>    storage_account_url = string,<br/>    schedule = object({<br/>      frequency_interval       = number,<br/>      frequency_unit           = string,<br/>      keep_at_least_one_backup = optional(bool),<br/>      retention_period_days    = optional(number)<br/>    })<br/>  })</pre> | `null` | no |
| <a name="input_client_affinity_enabled"></a> [client\_affinity\_enabled](#input\_client\_affinity\_enabled) | Should Client Affinity be enabled? | `bool` | `null` | no |
| <a name="input_client_certificate_enabled"></a> [client\_certificate\_enabled](#input\_client\_certificate\_enabled) | Should Client Certificates be enabled? | `bool` | `null` | no |
| <a name="input_client_certificate_exclusion_paths"></a> [client\_certificate\_exclusion\_paths](#input\_client\_certificate\_exclusion\_paths) | Paths to exclude when using client certificates, separated by ; | `string` | `null` | no |
| <a name="input_client_certificate_mode"></a> [client\_certificate\_mode](#input\_client\_certificate\_mode) | The Client Certificate mode. Possible values are `Required`, `Optional`, and `OptionalInteractiveUser`. This property has no effect when `client_certificate_enabled` is `false`. Defaults to `Required`. | `string` | `null` | no |
| <a name="input_connection_string"></a> [connection\_string](#input\_connection\_string) | One or more `connection_string` blocks as defined below. | <pre>set(object({<br/>    name  = string,<br/>    type  = string,<br/>    value = string<br/>  }))</pre> | `null` | no |
| <a name="input_datadog_api_key"></a> [datadog\_api\_key](#input\_datadog\_api\_key) | Datadog API key | `string` | n/a | yes |
| <a name="input_datadog_env"></a> [datadog\_env](#input\_datadog\_env) | Datadog Environment tag, used for Unified Service Tagging. | `string` | `null` | no |
| <a name="input_datadog_service"></a> [datadog\_service](#input\_datadog\_service) | Datadog Service tag, used for Unified Service Tagging. | `string` | `null` | no |
| <a name="input_datadog_site"></a> [datadog\_site](#input\_datadog\_site) | n/a | `string` | `"datadoghq.com"` | no |
| <a name="input_datadog_version"></a> [datadog\_version](#input\_datadog\_version) | Datadog Version tag, used for Unified Service Tagging. | `string` | `null` | no |
| <a name="input_enabled"></a> [enabled](#input\_enabled) | Should the Windows Web App be enabled? Defaults to `true`. | `bool` | `null` | no |
| <a name="input_ftp_publish_basic_authentication_enabled"></a> [ftp\_publish\_basic\_authentication\_enabled](#input\_ftp\_publish\_basic\_authentication\_enabled) | Should the default FTP Basic Authentication publishing profile be enabled. Defaults to `true`. | `bool` | `null` | no |
| <a name="input_https_only"></a> [https\_only](#input\_https\_only) | Should the Windows Web App require HTTPS connections. Defaults to `false`. | `bool` | `null` | no |
| <a name="input_identity"></a> [identity](#input\_identity) | n/a | <pre>object({<br/>    identity_ids = optional(set(string)),<br/>    type         = string<br/>  })</pre> | `null` | no |
| <a name="input_key_vault_reference_identity_id"></a> [key\_vault\_reference\_identity\_id](#input\_key\_vault\_reference\_identity\_id) | The User Assigned Identity ID used for accessing KeyVault secrets. The identity must be assigned to the application in the `identity` block. [For more information see - Access vaults with a user-assigned identity](https://docs.microsoft.com/azure/app-service/app-service-key-vault-references#access-vaults-with-a-user-assigned-identity) | `string` | `null` | no |
| <a name="input_location"></a> [location](#input\_location) | The Azure Region where the Windows Web App should exist. Changing this forces a new Windows Web App to be created. | `string` | n/a | yes |
| <a name="input_logs"></a> [logs](#input\_logs) | n/a | <pre>object({<br/>    detailed_error_messages = optional(bool),<br/>    failed_request_tracing  = optional(bool),<br/>    application_logs = optional(object({<br/>      file_system_level = string,<br/>      azure_blob_storage = optional(object({<br/>        level             = string,<br/>        retention_in_days = number,<br/>        sas_url           = string<br/>      }))<br/>    })),<br/>    http_logs = optional(object({<br/>      azure_blob_storage = optional(object({<br/>        retention_in_days = optional(number),<br/>        sas_url           = string<br/>      })),<br/>      file_system = optional(object({<br/>        retention_in_days = number,<br/>        retention_in_mb   = number<br/>      }))<br/>    }))<br/>  })</pre> | `null` | no |
| <a name="input_name"></a> [name](#input\_name) | The name which should be used for this Windows Web App. Changing this forces a new Windows Web App to be created. | `string` | n/a | yes |
| <a name="input_public_network_access_enabled"></a> [public\_network\_access\_enabled](#input\_public\_network\_access\_enabled) | Should public network access be enabled for the Web App. Defaults to `true`. | `bool` | `null` | no |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | The name of the Resource Group where the Windows Web App should exist. Changing this forces a new Windows Web App to be created. | `string` | n/a | yes |
| <a name="input_service_plan_id"></a> [service\_plan\_id](#input\_service\_plan\_id) | The ID of the Service Plan that this Windows App Service will be created in. | `string` | n/a | yes |
| <a name="input_site_config"></a> [site\_config](#input\_site\_config) | n/a | <pre>object({<br/>    always_on                                     = optional(bool),<br/>    api_definition_url                            = optional(string),<br/>    api_management_api_id                         = optional(string),<br/>    app_command_line                              = optional(string),<br/>    container_registry_managed_identity_client_id = optional(string),<br/>    container_registry_use_managed_identity       = optional(bool),<br/>    ftps_state                                    = optional(string),<br/>    health_check_eviction_time_in_min             = optional(number),<br/>    health_check_path                             = optional(string),<br/>    http2_enabled                                 = optional(bool),<br/>    ip_restriction_default_action                 = optional(string),<br/>    load_balancing_mode                           = optional(string),<br/>    local_mysql_enabled                           = optional(bool),<br/>    managed_pipeline_mode                         = optional(string),<br/>    minimum_tls_version                           = optional(string),<br/>    remote_debugging_enabled                      = optional(bool),<br/>    scm_ip_restriction_default_action             = optional(string),<br/>    scm_minimum_tls_version                       = optional(string),<br/>    scm_use_main_ip_restriction                   = optional(bool),<br/>    use_32_bit_worker                             = optional(bool),<br/>    vnet_route_all_enabled                        = optional(bool),<br/>    websockets_enabled                            = optional(bool),<br/>    application_stack = optional(object({<br/>      current_stack  = optional(string),<br/>      dotnet_version = optional(string),<br/>      java_version   = optional(string),<br/>      node_version   = optional(string)<br/>    })),<br/>    auto_heal_setting = optional(object({<br/>      action = object({<br/>        action_type = string,<br/>        custom_action = optional(object({<br/>          executable = string,<br/>          parameters = optional(string)<br/>        }))<br/>      }),<br/>      trigger = object({<br/>        private_memory_kb = optional(number),<br/>        requests = optional(object({<br/>          count    = number,<br/>          interval = string<br/>        })),<br/>        slow_request = optional(object({<br/>          count      = number,<br/>          interval   = string,<br/>          time_taken = string<br/>        })),<br/>        slow_request_with_path = optional(list(object({<br/>          count      = number,<br/>          interval   = string,<br/>          path       = optional(string),<br/>          time_taken = string<br/>        }))),<br/>        status_code = optional(set(object({<br/>          count             = number,<br/>          interval          = string,<br/>          path              = optional(string),<br/>          status_code_range = string,<br/>          sub_status        = optional(number),<br/>          win32_status_code = optional(number)<br/>        })))<br/>      })<br/>    })),<br/>    cors = optional(object({<br/>      allowed_origins     = optional(set(string)),<br/>      support_credentials = optional(bool)<br/>    })),<br/>    handler_mapping = optional(set(object({<br/>      arguments             = optional(string),<br/>      extension             = string,<br/>      script_processor_path = string<br/>    }))),<br/>    ip_restriction = optional(list(object({<br/>      action      = optional(string),<br/>      description = optional(string),<br/>      headers = optional(list(object({<br/>        x_azure_fdid      = list(string),<br/>        x_fd_health_probe = list(string),<br/>        x_forwarded_for   = list(string),<br/>        x_forwarded_host  = list(string)<br/>      }))),<br/>      ip_address                = optional(string),<br/>      priority                  = optional(number),<br/>      service_tag               = optional(string),<br/>      virtual_network_subnet_id = optional(string)<br/>    }))),<br/>    scm_ip_restriction = optional(list(object({<br/>      action      = optional(string),<br/>      description = optional(string),<br/>      headers = optional(list(object({<br/>        x_azure_fdid      = list(string),<br/>        x_fd_health_probe = list(string),<br/>        x_forwarded_for   = list(string),<br/>        x_forwarded_host  = list(string)<br/>      }))),<br/>      ip_address                = optional(string),<br/>      priority                  = optional(number),<br/>      service_tag               = optional(string),<br/>      virtual_network_subnet_id = optional(string)<br/>    }))),<br/>    virtual_application = optional(set(object({<br/>      physical_path = string,<br/>      preload       = bool,<br/>      virtual_path  = string,<br/>      virtual_directory = optional(set(object({<br/>        physical_path = optional(string),<br/>        virtual_path  = optional(string)<br/>      })))<br/>    })))<br/>  })</pre> | n/a | yes |
| <a name="input_sticky_settings"></a> [sticky\_settings](#input\_sticky\_settings) | n/a | <pre>object({<br/>    app_setting_names       = optional(list(string)),<br/>    connection_string_names = optional(list(string))<br/>  })</pre> | `null` | no |
| <a name="input_storage_account"></a> [storage\_account](#input\_storage\_account) | One or more `storage_account` blocks as defined below. | <pre>set(object({<br/>    access_key   = string,<br/>    account_name = string,<br/>    mount_path   = optional(string),<br/>    name         = string,<br/>    share_name   = string,<br/>    type         = string<br/>  }))</pre> | `null` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | A mapping of tags which should be assigned to the Windows Web App. | `map(string)` | `null` | no |
| <a name="input_timeouts"></a> [timeouts](#input\_timeouts) | n/a | <pre>object({<br/>    create = optional(string),<br/>    delete = optional(string),<br/>    read   = optional(string),<br/>    update = optional(string)<br/>  })</pre> | `null` | no |
| <a name="input_virtual_network_backup_restore_enabled"></a> [virtual\_network\_backup\_restore\_enabled](#input\_virtual\_network\_backup\_restore\_enabled) | Whether backup and restore operations over the linked virtual network are enabled. Defaults to `false`. | `bool` | `null` | no |
| <a name="input_virtual_network_subnet_id"></a> [virtual\_network\_subnet\_id](#input\_virtual\_network\_subnet\_id) | The subnet id which will be used by this Web App for [regional virtual network integration](https://docs.microsoft.com/en-us/azure/app-service/overview-vnet-integration#regional-virtual-network-integration). | `string` | `null` | no |
| <a name="input_webdeploy_publish_basic_authentication_enabled"></a> [webdeploy\_publish\_basic\_authentication\_enabled](#input\_webdeploy\_publish\_basic\_authentication\_enabled) | Should the default WebDeploy Basic Authentication publishing credentials enabled. Defaults to `true`. | `bool` | `null` | no |
| <a name="input_zip_deploy_file"></a> [zip\_deploy\_file](#input\_zip\_deploy\_file) | The local path and filename of the Zip packaged application to deploy to this Windows Web App. **Note:** Using this value requires either `WEBSITE_RUN_FROM_PACKAGE=1` or `SCM_DO_BUILD_DURING_DEPLOYMENT=true` to be set on the App in `app_settings`. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_app_settings"></a> [app\_settings](#output\_app\_settings) | A map of key-value pairs of App Settings. |
| <a name="output_auth_settings"></a> [auth\_settings](#output\_auth\_settings) | n/a |
| <a name="output_auth_settings_v2"></a> [auth\_settings\_v2](#output\_auth\_settings\_v2) | n/a |
| <a name="output_backup"></a> [backup](#output\_backup) | n/a |
| <a name="output_client_affinity_enabled"></a> [client\_affinity\_enabled](#output\_client\_affinity\_enabled) | Should Client Affinity be enabled? |
| <a name="output_client_certificate_enabled"></a> [client\_certificate\_enabled](#output\_client\_certificate\_enabled) | Should Client Certificates be enabled? |
| <a name="output_client_certificate_exclusion_paths"></a> [client\_certificate\_exclusion\_paths](#output\_client\_certificate\_exclusion\_paths) | Paths to exclude when using client certificates, separated by ; |
| <a name="output_client_certificate_mode"></a> [client\_certificate\_mode](#output\_client\_certificate\_mode) | The Client Certificate mode. Possible values are `Required`, `Optional`, and `OptionalInteractiveUser`. This property has no effect when `client_certificate_enabled` is `false`. Defaults to `Required`. |
| <a name="output_connection_string"></a> [connection\_string](#output\_connection\_string) | One or more `connection_string` blocks as defined below. |
| <a name="output_custom_domain_verification_id"></a> [custom\_domain\_verification\_id](#output\_custom\_domain\_verification\_id) | The identifier used by App Service to perform domain ownership verification via DNS TXT record. |
| <a name="output_default_hostname"></a> [default\_hostname](#output\_default\_hostname) | The default hostname of the Windows Web App. |
| <a name="output_enabled"></a> [enabled](#output\_enabled) | Should the Windows Web App be enabled? Defaults to `true`. |
| <a name="output_ftp_publish_basic_authentication_enabled"></a> [ftp\_publish\_basic\_authentication\_enabled](#output\_ftp\_publish\_basic\_authentication\_enabled) | Should the default FTP Basic Authentication publishing profile be enabled. Defaults to `true`. |
| <a name="output_hosting_environment_id"></a> [hosting\_environment\_id](#output\_hosting\_environment\_id) | The ID of the App Service Environment used by App Service. |
| <a name="output_https_only"></a> [https\_only](#output\_https\_only) | Should the Windows Web App require HTTPS connections. Defaults to `false`. |
| <a name="output_id"></a> [id](#output\_id) | The ID of the Windows Web App. |
| <a name="output_identity"></a> [identity](#output\_identity) | n/a |
| <a name="output_key_vault_reference_identity_id"></a> [key\_vault\_reference\_identity\_id](#output\_key\_vault\_reference\_identity\_id) | The User Assigned Identity ID used for accessing KeyVault secrets. The identity must be assigned to the application in the `identity` block. [For more information see - Access vaults with a user-assigned identity](https://docs.microsoft.com/azure/app-service/app-service-key-vault-references#access-vaults-with-a-user-assigned-identity) |
| <a name="output_kind"></a> [kind](#output\_kind) | The Kind value for this Windows Web App. |
| <a name="output_location"></a> [location](#output\_location) | The Azure Region where the Windows Web App should exist. Changing this forces a new Windows Web App to be created. |
| <a name="output_logs"></a> [logs](#output\_logs) | n/a |
| <a name="output_name"></a> [name](#output\_name) | The name which should be used for this Windows Web App. Changing this forces a new Windows Web App to be created. |
| <a name="output_outbound_ip_address_list"></a> [outbound\_ip\_address\_list](#output\_outbound\_ip\_address\_list) | A list of outbound IP addresses - such as `["52.23.25.3", "52.143.43.12"]` |
| <a name="output_outbound_ip_addresses"></a> [outbound\_ip\_addresses](#output\_outbound\_ip\_addresses) | A comma separated list of outbound IP addresses - such as `52.23.25.3,52.143.43.12`. |
| <a name="output_possible_outbound_ip_address_list"></a> [possible\_outbound\_ip\_address\_list](#output\_possible\_outbound\_ip\_address\_list) | A list of possible outbound ip address. |
| <a name="output_possible_outbound_ip_addresses"></a> [possible\_outbound\_ip\_addresses](#output\_possible\_outbound\_ip\_addresses) | A comma separated list of outbound IP addresses - such as `52.23.25.3,52.143.43.12,52.143.43.17` - not all of which are necessarily in use. Superset of `outbound_ip_addresses`. |
| <a name="output_public_network_access_enabled"></a> [public\_network\_access\_enabled](#output\_public\_network\_access\_enabled) | Should public network access be enabled for the Web App. Defaults to `true`. |
| <a name="output_resource_group_name"></a> [resource\_group\_name](#output\_resource\_group\_name) | The name of the Resource Group where the Windows Web App should exist. Changing this forces a new Windows Web App to be created. |
| <a name="output_service_plan_id"></a> [service\_plan\_id](#output\_service\_plan\_id) | The ID of the Service Plan that this Windows App Service will be created in. |
| <a name="output_site_config"></a> [site\_config](#output\_site\_config) | n/a |
| <a name="output_site_credential"></a> [site\_credential](#output\_site\_credential) | n/a |
| <a name="output_sticky_settings"></a> [sticky\_settings](#output\_sticky\_settings) | n/a |
| <a name="output_storage_account"></a> [storage\_account](#output\_storage\_account) | One or more `storage_account` blocks as defined below. |
| <a name="output_tags"></a> [tags](#output\_tags) | A mapping of tags which should be assigned to the Windows Web App. |
| <a name="output_timeouts"></a> [timeouts](#output\_timeouts) | n/a |
| <a name="output_virtual_network_backup_restore_enabled"></a> [virtual\_network\_backup\_restore\_enabled](#output\_virtual\_network\_backup\_restore\_enabled) | Whether backup and restore operations over the linked virtual network are enabled. Defaults to `false`. |
| <a name="output_virtual_network_image_pull_enabled"></a> [virtual\_network\_image\_pull\_enabled](#output\_virtual\_network\_image\_pull\_enabled) | Whether traffic for the image pull should be routed over the virtual network. |
| <a name="output_virtual_network_subnet_id"></a> [virtual\_network\_subnet\_id](#output\_virtual\_network\_subnet\_id) | The subnet id which will be used by this Web App for [regional virtual network integration](https://docs.microsoft.com/en-us/azure/app-service/overview-vnet-integration#regional-virtual-network-integration). |
| <a name="output_webdeploy_publish_basic_authentication_enabled"></a> [webdeploy\_publish\_basic\_authentication\_enabled](#output\_webdeploy\_publish\_basic\_authentication\_enabled) | Should the default WebDeploy Basic Authentication publishing credentials enabled. Defaults to `true`. |
| <a name="output_zip_deploy_file"></a> [zip\_deploy\_file](#output\_zip\_deploy\_file) | The local path and filename of the Zip packaged application to deploy to this Windows Web App. **Note:** Using this value requires either `WEBSITE_RUN_FROM_PACKAGE=1` or `SCM_DO_BUILD_DURING_DEPLOYMENT=true` to be set on the App in `app_settings`. |
<!-- END_TF_DOCS -->
