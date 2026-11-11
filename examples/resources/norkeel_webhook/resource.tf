variable "webhook_secret_rotation" {
  type    = string
  default = "2026-09"
}

# The signing secret is returned only on create and on rotation; change
# rotate_secret to rotate it. At most 10 webhooks per organization.
resource "norkeel_webhook" "alerts" {
  url           = "https://hooks.example.com/norkeel"
  description   = "Cluster lifecycle and budget alerts"
  events        = ["cluster.ready", "cluster.degraded", "usage.alert"]
  rotate_secret = var.webhook_secret_rotation
}

output "webhook_secret" {
  value     = norkeel_webhook.alerts.secret
  sensitive = true
}
