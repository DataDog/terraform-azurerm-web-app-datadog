variable "datadog_api_key" {
  type      = string
  sensitive = true
  nullable  = false
}

variable "datadog_site" {
  type    = string
  default = "datadoghq.com"
}

variable "subscription_id" {
  type     = string
  nullable = false
}

variable "resource_group_name" {
  type     = string
  nullable = false
}

variable "name" {
  type     = string
  nullable = false
}

variable "location" {
  type     = string
  nullable = false
}

