provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

# Ephemeral, uniquely named workload host. The resource group carries the same
# name prefix + freshness tag as the web app so the sweeper can reap a leaked
# group wholesale.
resource "azurerm_resource_group" "this" {
  name     = var.resource_group_name
  location = var.location
  tags     = var.tags
}

resource "azurerm_service_plan" "this" {
  name                = "${var.name}-plan"
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location
  os_type             = "Linux"
  sku_name            = "P1v2"
  tags                = var.tags
}

# The module under test. Wrapping the web app is the instrumentation mechanism:
# `terraform apply` is APPLY, `terraform destroy` is REMOVE.
module "datadog_linux_web_app" {
  source = "../../../modules/linux"

  datadog_api_key = var.datadog_api_key
  datadog_site    = var.datadog_site
  datadog_service = var.datadog_service
  datadog_env     = var.datadog_env
  datadog_version = var.datadog_version
  sidecar_image   = var.sidecar_image

  name                = var.name
  location            = var.location
  resource_group_name = azurerm_resource_group.this.name
  service_plan_id     = azurerm_service_plan.this.id
  https_only          = true

  site_config = {
    # The prebuilt workload package needs an explicit startup command to boot on
    # a Linux Web App (the published self-monitoring template sets the same).
    # This is workload startup config, not part of the module's instrumentation.
    app_command_line = "npm start"
    application_stack = {
      node_version = "22-lts"
    }
  }

  tags = var.tags
}
