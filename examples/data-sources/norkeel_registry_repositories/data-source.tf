# Repositories of the organization's registry, read live from the registry.
data "norkeel_registry_repositories" "all" {}

output "repositories" {
  value = { for r in data.norkeel_registry_repositories.all.repositories : r.name => r.artifact_count }
}
