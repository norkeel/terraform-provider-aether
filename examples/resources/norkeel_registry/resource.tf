# Destroying the registry disables it and deletes EVERY robot account of the organization;
# images are kept unless purge_on_destroy = true (irreversible).
resource "norkeel_registry" "this" {
  purge_on_destroy = false
}

# depends_on makes Terraform destroy the robot account before the registry.
resource "norkeel_registry_robot_account" "ci" {
  name        = "ci"
  permissions = ["push", "pull"]

  depends_on = [norkeel_registry.this]
}

# An enabled nightly/weekly "delete" policy runs within about a minute of the apply:
# start with action = "protect" or enabled = false.
resource "norkeel_registry_cleanup_policy" "old_untagged" {
  name     = "old-untagged"
  action   = "protect"
  schedule = "nightly"
  enabled  = false

  conditions = [
    { type = "untagged", value = "true" },
    { type = "age", op = "gt", value = "30d" },
  ]

  depends_on = [norkeel_registry.this]
}

output "robot_secret" {
  value     = norkeel_registry_robot_account.ci.secret
  sensitive = true
}
