# Unless explicitly stated otherwise all files in this repository are licensed under the Apache-2.0 License.
# This product includes software developed at Datadog (https://www.datadoghq.com/) Copyright 2025 Datadog, Inc.

# Local Definitions
locals {
  module_version      = "1.0.1"
  is_container        = try(var.site_config.application_stack.docker_image_name, null) != null
  is_dotnet_container = coalesce(try(var.container_config.is_dotnet, null), false)
  is_musl_container   = coalesce(try(var.container_config.is_musl, null), false)

  datadog_service = coalesce(var.datadog_service, var.name)
}


check "container_has_config" {
  assert {
    condition     = !(local.is_container && var.container_config == null)
    error_message = "The 'container_config' variable must be set if the application is a container."
  }
}


# Sidecar Logic/Installation
locals {
  sidecar_container_name = "datadog-sidecar"
  sidecar_image          = "index.docker.io/datadog/serverless-init:latest"
  sidecar_port           = "8126"
}

# Convert to Sitecontainers if the customer is using a containerized app
# workaround: https://github.com/hashicorp/terraform-provider-azurerm/issues/25167#issuecomment-2586756232
resource "azapi_update_resource" "enable_sidecar" {
  count       = local.is_container ? 1 : 0 # only run if the app is containerized
  type        = "Microsoft.Web/sites@2022-03-01"
  resource_id = azurerm_linux_web_app.this.id
  body        = { properties = { siteConfig = { linuxFxVersion = "SITECONTAINERS" } } }
  lifecycle { replace_triggered_by = [azurerm_linux_web_app.this] }
}

locals {
  main_container_host  = try(trimprefix(trimprefix(var.site_config.application_stack.docker_registry_url, "https://"), "http://"), "")
  main_container_image = try("${local.main_container_host}/${var.site_config.application_stack.docker_image_name}", "")
}
resource "azapi_resource" "main_container" {
  count      = local.is_container ? 1 : 0 # only run if the app is containerized
  depends_on = [azapi_update_resource.enable_sidecar]
  type       = "Microsoft.Web/sites/sitecontainers@2024-11-01"
  parent_id  = azurerm_linux_web_app.this.id
  name       = "main"
  # https://learn.microsoft.com/en-us/rest/api/appservice/web-apps/create-or-update-site-container?view=rest-appservice-2024-11-01#request-body
  body = {
    properties = {
      image          = local.main_container_image
      isMain         = true
      authType       = var.site_config.application_stack.docker_registry_username != null ? "UserCredentials" : "Anonymous"
      userName       = var.site_config.application_stack.docker_registry_username
      passwordSecret = var.site_config.application_stack.docker_registry_password
      targetPort     = try(var.container_config.port, "8080")
    }
  }
}


# Resource Implementation Locals
locals {
  is_dotnet = local.is_dotnet_container || var.site_config.application_stack.dotnet_version != null

  app_settings = merge(
    {
      DD_API_KEY = var.datadog_api_key
      DD_SITE    = var.datadog_site
      DD_SERVICE = local.datadog_service,
    },
    var.datadog_env != null ? { DD_ENV = var.datadog_env } : {},
    var.datadog_version != null ? { DD_VERSION = var.datadog_version } : {},
    local.is_dotnet ? {
      DD_DOTNET_TRACER_HOME    = "/home/site/wwwroot/datadog",
      DD_TRACE_LOG_DIRECTORY   = "/home/LogFiles/dotnet",
      CORECLR_ENABLE_PROFILING = "1",
      CORECLR_PROFILER         = "{846F5F1C-F9AE-4B07-969E-05C26BC060D8}",
      CORECLR_PROFILER_PATH = (
        local.is_musl_container
        ? "/home/site/wwwroot/datadog/linux-musl-x64/Datadog.Trace.ClrProfiler.Native.so"
        : "/home/site/wwwroot/datadog/linux-x64/Datadog.Trace.ClrProfiler.Native.so"
      ),
    } : {},
    var.app_settings
  )
  tags = merge(
    { service = local.datadog_service, dd_sls_terraform_module = local.module_version },
    var.datadog_env != null ? { env = var.datadog_env } : {},
    var.datadog_version != null ? { version = var.datadog_version } : {},
    var.tags
  )
}

resource "azapi_resource" "datadog_sidecar" {
  type      = "Microsoft.Web/sites/sitecontainers@2024-11-01"
  parent_id = azurerm_linux_web_app.this.id
  name      = local.sidecar_container_name
  body = { properties = {
    image      = local.sidecar_image
    isMain     = false
    targetPort = local.sidecar_port
    environmentVariables = [for k, _ in local.app_settings : {
      name  = k
      value = k
    }]
  } }
}
