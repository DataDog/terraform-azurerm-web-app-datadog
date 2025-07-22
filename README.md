# Datadog Terraform module for Azure Linux Web Apps

Use this Terraform module to install Datadog Serverless Monitoring for Azure Linux Web Apps.

This Terraform module wraps the [azurerm_linux_web_app resource](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/linux_web_app) and automatically configures your Cloud Run application for Datadog Serverless Monitoring by:

* creating the `azurerm_linux_web_app` resource invocation with the proper additional environment variables
* enabling the Datadog agent as a sidecar container to collect metrics, traces, and logs