provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

resource "azurerm_resource_group" "example" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_service_plan" "example" {
  resource_group_name = azurerm_resource_group.example.name
  name                = "${var.name}-service-plan"
  location            = var.location
  sku_name            = "P1v2"
  os_type             = "Linux"
}

module "datadog_linux_web_app" {
  source          = "../../"
  datadog_api_key = var.datadog_api_key
  datadog_site    = var.datadog_site
  datadog_env     = "dev"
  datadog_service = "my-service"
  datadog_version = "1.0.0"

  resource_group_name = azurerm_resource_group.example.name
  name                = var.name
  location            = var.location
  service_plan_id     = azurerm_service_plan.example.id
  site_config = {
    application_stack = {
        java_version = "21"
        java_server = "JAVA"
        java_server_version = "21"
    }
  }
  app_settings = { # additional app settings/features
    DD_PROFILING_ENABLED = "true" # example feature enablement
  }
  tags = { # additional resource tags
    test = "true"
  }
}

resource "azurerm_app_service_source_control" "code_deployment" {
  app_id                 = module.datadog_linux_web_app.id
  repo_url               = "https://github.com/Azure-Samples/java-docs-hello-world"
  branch                 = "main"
  use_manual_integration = true
  use_mercurial          = false
}
