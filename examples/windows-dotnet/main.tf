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
  os_type             = "Windows"
}

module "datadog_windows_web_app" {
  source          = "../../modules/windows"
  datadog_api_key = var.datadog_api_key
  datadog_site    = var.datadog_site
  datadog_env     = "dev"
  datadog_service = "my-service"
  datadog_version = "1.0.0"

  resource_group_name = azurerm_resource_group.example.name
  name                = var.name
  location            = var.location
  service_plan_id     = azurerm_service_plan.example.id
  https_only          = true
  site_config = {
    application_stack = {
      dotnet_version = "v9.0"
    }
  }
  app_settings = { # additional app settings/features

    DD_PROFILING_ENABLED = "true" # example feature enablement

    SCM_DO_BUILD_DURING_DEPLOYMENT = "true" # Required for local deployment below
  }
  tags = { # additional resource tags
    test = "true"
  }
}

resource "terraform_data" "code_deployment" { # Basic local deployment setup, replace with your actual deployment method in prod
  depends_on = [module.datadog_windows_web_app]
  provisioner "local-exec" {
    command = <<EOT
    cd src
    zip -r code.zip Properties appsettings.json src.csproj Program.cs
    az webapp deploy -g ${azurerm_resource_group.example.name} -n ${module.datadog_windows_web_app.name} --src-path code.zip --type zip
    EOT
  }
}
