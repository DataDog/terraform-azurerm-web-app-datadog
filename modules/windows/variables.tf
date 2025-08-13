# Unless explicitly stated otherwise all files in this repository are licensed under the Apache-2.0 License.
# This product includes software developed at Datadog (https://www.datadoghq.com/) Copyright 2025 Datadog, Inc.

variable "datadog_api_key" {
  type        = string
  description = "Datadog API key"
  sensitive   = true
}

variable "datadog_site" {
  type    = string
  default = "datadoghq.com"
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
