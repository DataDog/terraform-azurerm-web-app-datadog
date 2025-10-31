provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

resource "azurerm_resource_group" "example" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_container_registry" "example" {
  name                   = "${replace(var.name, "/\\W|_|\\s/", "")}acr"
  resource_group_name    = azurerm_resource_group.example.name
  location               = var.location
  sku                    = "Standard"
  anonymous_pull_enabled = true
}

resource "terraform_data" "acr_push_image" {
  provisioner "local-exec" {
    command = <<EOT
        az acr login --name ${azurerm_container_registry.example.name}
        docker buildx build --platform linux/amd64 -t ${azurerm_container_registry.example.login_server}/hello-world:latest --push ./src
    EOT
  }
  depends_on = [azurerm_container_registry.example]
}

resource "azurerm_service_plan" "example" {
  resource_group_name = azurerm_resource_group.example.name
  name                = "${var.name}-service-plan"
  location            = var.location
  sku_name            = "P1v2"
  os_type             = "Linux"
}

module "datadog_linux_web_app" {
  source          = "../../modules/linux"
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
      docker_registry_url = "https://${azurerm_container_registry.example.login_server}"
      docker_image_name   = "hello-world:latest"
    }
  }
  container_config = {
    port = "8080"
  }
  app_settings = {                # additional app settings/features
    DD_PROFILING_ENABLED = "true" # example feature enablement
  }
  tags = { # additional resource tags
    test = "true"
  }
}
