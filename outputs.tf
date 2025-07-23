output "app_settings" {
  value = azurerm_linux_web_app.this.app_settings
}

output "client_affinity_enabled" {
  value = azurerm_linux_web_app.this.client_affinity_enabled
}

output "client_certificate_enabled" {
  value = azurerm_linux_web_app.this.client_certificate_enabled
}

output "client_certificate_exclusion_paths" {
  description = <<DESCRIPTION
Paths to exclude when using client certificates, separated by ;
DESCRIPTION
  value       = azurerm_linux_web_app.this.client_certificate_exclusion_paths
}

output "client_certificate_mode" {
  value = azurerm_linux_web_app.this.client_certificate_mode
}

output "custom_domain_verification_id" {
  value = azurerm_linux_web_app.this.custom_domain_verification_id
}

output "default_hostname" {
  value = azurerm_linux_web_app.this.default_hostname
}

output "enabled" {
  value = azurerm_linux_web_app.this.enabled
}

output "ftp_publish_basic_authentication_enabled" {
  value = azurerm_linux_web_app.this.ftp_publish_basic_authentication_enabled
}

output "hosting_environment_id" {
  value = azurerm_linux_web_app.this.hosting_environment_id
}

output "https_only" {
  value = azurerm_linux_web_app.this.https_only
}

output "id" {
  value = azurerm_linux_web_app.this.id
}

output "key_vault_reference_identity_id" {
  value = azurerm_linux_web_app.this.key_vault_reference_identity_id
}

output "kind" {
  value = azurerm_linux_web_app.this.kind
}

output "location" {
  value = azurerm_linux_web_app.this.location
}

output "name" {
  value = azurerm_linux_web_app.this.name
}

output "outbound_ip_address_list" {
  value = azurerm_linux_web_app.this.outbound_ip_address_list
}

output "outbound_ip_addresses" {
  value = azurerm_linux_web_app.this.outbound_ip_addresses
}

output "possible_outbound_ip_address_list" {
  value = azurerm_linux_web_app.this.possible_outbound_ip_address_list
}

output "possible_outbound_ip_addresses" {
  value = azurerm_linux_web_app.this.possible_outbound_ip_addresses
}

output "public_network_access_enabled" {
  value = azurerm_linux_web_app.this.public_network_access_enabled
}

output "resource_group_name" {
  value = azurerm_linux_web_app.this.resource_group_name
}

output "service_plan_id" {
  value = azurerm_linux_web_app.this.service_plan_id
}

output "site_credential" {
  value = azurerm_linux_web_app.this.site_credential
}

output "tags" {
  value = azurerm_linux_web_app.this.tags
}

output "virtual_network_backup_restore_enabled" {
  value = azurerm_linux_web_app.this.virtual_network_backup_restore_enabled
}

output "virtual_network_subnet_id" {
  value = azurerm_linux_web_app.this.virtual_network_subnet_id
}

output "vnet_image_pull_enabled" {
  value = azurerm_linux_web_app.this.vnet_image_pull_enabled
}

output "webdeploy_publish_basic_authentication_enabled" {
  value = azurerm_linux_web_app.this.webdeploy_publish_basic_authentication_enabled
}

output "zip_deploy_file" {
  description = <<DESCRIPTION
The local path and filename of the Zip packaged application to deploy to this Linux Web App. **Note:** Using this value requires either `WEBSITE_RUN_FROM_PACKAGE=1` or `SCM_DO_BUILD_DURING_DEPLOYMENT=true` to be set on the App in `app_settings`.
DESCRIPTION
  value       = azurerm_linux_web_app.this.zip_deploy_file
}

output "auth_settings" {
  value = azurerm_linux_web_app.this.auth_settings
}

output "auth_settings_v2" {
  value = azurerm_linux_web_app.this.auth_settings_v2
}

output "backup" {
  value = azurerm_linux_web_app.this.backup
}

output "connection_string" {
  value = azurerm_linux_web_app.this.connection_string
}

output "identity" {
  value = azurerm_linux_web_app.this.identity
}

output "logs" {
  value = azurerm_linux_web_app.this.logs
}

output "site_config" {
  value = azurerm_linux_web_app.this.site_config
}

output "sticky_settings" {
  value = azurerm_linux_web_app.this.sticky_settings
}

output "storage_account" {
  value = azurerm_linux_web_app.this.storage_account
}

output "timeouts" {
  value = azurerm_linux_web_app.this.timeouts
}