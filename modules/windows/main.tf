# Resource Implementation Locals
locals {
  app_settings = merge(
    {
      DD_API_KEY = var.datadog_api_key
      DD_SITE    = var.datadog_site
      DD_ENV     = var.datadog_env,
      DD_SERVICE = var.datadog_service,
      DD_VERSION = var.datadog_version,
    },
    var.app_settings
  )
  tags = merge(
    {
      env     = var.datadog_env,
      service = var.datadog_service,
      version = var.datadog_version,
    },
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
