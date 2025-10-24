# Unless explicitly stated otherwise all files in this repository are licensed under the Apache-2.0 License.
# This product includes software developed at Datadog (https://www.datadoghq.com/) Copyright 2025 Datadog, Inc.

# Local Definitions
locals {
  datadog_service = coalesce(var.datadog_service, var.name)
}

# Resource Implementation Locals
locals {
  app_settings = merge(
    {
      DD_API_KEY = var.datadog_api_key
      DD_SITE    = var.datadog_site
      DD_SERVICE = local.datadog_service,
    },
    var.datadog_env != null ? { DD_ENV = var.datadog_env } : {},
    var.datadog_version != null ? { DD_VERSION = var.datadog_version } : {},
    var.app_settings
  )
  tags = merge(
    { service = local.datadog_service },
    var.datadog_env != null ? { env = var.datadog_env } : {},
    var.datadog_version != null ? { version = var.datadog_version } : {},
    var.tags
  )

}

# Extension Logic/Installation
locals {
  datadog_extension_name = (
    var.site_config.application_stack.current_stack == "dotnet" || var.site_config.application_stack.dotnet_version != null ? "Datadog.AzureAppServices.DotNet" :
    var.site_config.application_stack.current_stack == "java" || var.site_config.application_stack.java_version != null ? "Datadog.AzureAppServices.Java.Apm" :
    var.site_config.application_stack.current_stack == "node" || var.site_config.application_stack.node_version != null ? "Datadog.AzureAppServices.Node.Apm" : null
  )
}
check "valid_runtime" {
  assert {
    condition     = local.datadog_extension_name != null
    error_message = "Datadog extension is not supported for the specified application stack."
  }
}

resource "azapi_resource" "datadog_extension" {
  type      = "Microsoft.Web/sites/siteextensions@2024-11-01"
  parent_id = azurerm_windows_web_app.this.id
  name      = local.datadog_extension_name
  location  = var.location
}
