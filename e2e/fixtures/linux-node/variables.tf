variable "subscription_id" {
  type     = string
  nullable = false
}

variable "datadog_api_key" {
  type      = string
  sensitive = true
  nullable  = false
}

variable "datadog_site" {
  type    = string
  default = "datadoghq.com"
}

# Unified Service Tagging values. The e2e harness sets datadog_service to the
# run-unique resource name so telemetry can be filtered by run id, and asserts
# these exact values on ingested traces/logs (identity, not existence).
variable "datadog_service" {
  type     = string
  nullable = false
}

variable "datadog_env" {
  type    = string
  default = "e2e"
}

variable "datadog_version" {
  type    = string
  default = "1.0.0"
}

variable "name" {
  type     = string
  nullable = false
}

variable "location" {
  type    = string
  default = "eastus2"
}

variable "resource_group_name" {
  type     = string
  nullable = false
}

# Pinned sidecar image. The default pins serverless-init so a telemetry failure
# blames the module wiring, not an upstream agent regression. CI overrides this
# with the pinned tag from the e2e workflow.
variable "sidecar_image" {
  type    = string
  default = "index.docker.io/datadog/serverless-init:latest"
}

# Resource tags applied via the module. The harness injects the freshness tag
# (one_e2e_created:<unix-ts>) and the run-id marker here so the cross-repo
# sweeper can identify and reap leaked resources.
variable "tags" {
  type    = map(string)
  default = {}
}
