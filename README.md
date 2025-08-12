# Datadog Terraform modules for Azure Web Apps

Use these Terraform modules to install Datadog Serverless Monitoring for Azure Linux & Windows Web Apps.

These Terraform modules wrap the following resources:
- [azurerm_linux_web_app](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/linux_web_app)
- [azurerm_windows_web_app](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/windows_web_app)

And automatically configure your Web App for Datadog Serverless Monitoring by:
- creating the `azurerm_linux_web_app` resource invocation with the proper additional environment variables
- enabling the Datadog agent as a sidecar container to collect metrics, traces, and logs