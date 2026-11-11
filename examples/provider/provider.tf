terraform {
  required_version = ">= 1.5" # OpenTofu >= 1.6 works the same way
  required_providers {
    norkeel = {
      source  = "norkeel/norkeel"
      version = "~> 0.2"
    }
  }
}

variable "norkeel_api_token" {
  type      = string
  sensitive = true
  default   = null # unset here so the provider falls back to NORKEEL_API_TOKEN
}

# Create an API token in the Norkeel portal first (Settings → API Tokens). Role: Admin (the default).
provider "norkeel" {
  api_token = var.norkeel_api_token # or NORKEEL_API_TOKEN
  # endpoint defaults to https://api.aetherplatform.cloud (production);
  # use https://dev.api.aetherplatform.cloud for the dev environment (or NORKEEL_ENDPOINT).
}
