variable "datadog_api_key" {
  type        = string
  description = "Datadog API key"
  sensitive   = true
}

variable "datadog_site" {
  type     = string
  default  = "datadoghq.com"
  nullable = false
  validation {
    condition = contains(
      [
        "datadoghq.com",
        "datadoghq.eu",
        "us5.datadoghq.com",
        "us3.datadoghq.com",
        "ddog-gov.com",
        "ap1.datadoghq.com",
        "ap2.datadoghq.com",
      ],
    var.datadog_site)
    error_message = "Invalid Datadog site. Valid options are: 'datadoghq.com', 'datadoghq.eu', 'us5.datadoghq.com', 'us3.datadoghq.com', 'ddog-gov.com', 'ap1.datadoghq.com', or 'ap2.datadoghq.com'."
  }
}

variable "datadog_env" {
  type        = string
  nullable    = true
  default     = null
  description = "Datadog Environment tag, used for Unified Service Tagging."
}

variable "datadog_service" {
  type        = string
  nullable    = true
  default     = null
  description = "Datadog Service tag, used for Unified Service Tagging."
}

variable "datadog_version" {
  type        = string
  nullable    = true
  default     = null
  description = "Datadog Version tag, used for Unified Service Tagging."
}

variable "container_config" {
  type = object({
    port      = string
    is_dotnet = optional(bool)
    is_musl   = optional(bool)
  })
  description = "Additional Configuration for containerized applications. This is required if the application is a container."
  default     = null
  validation {
    condition     = var.container_config == null ? true : !(coalesce(var.container_config.is_musl, false) && !coalesce(var.container_config.is_dotnet, false))
    error_message = "The 'container_config.is_musl' variable can only be set to true if 'container_config.is_dotnet' is also true."
  }
}
