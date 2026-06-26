output "web_app_name" {
  value = module.datadog_linux_web_app.name
}

output "resource_group_name" {
  value = azurerm_resource_group.this.name
}

output "default_hostname" {
  value = module.datadog_linux_web_app.default_hostname
}
