# Offered Kubernetes versions; latest is null when the version window is unavailable.
data "norkeel_k8s_versions" "offered" {}

output "latest" {
  value = data.norkeel_k8s_versions.offered.latest
}

output "all" {
  value = [for v in data.norkeel_k8s_versions.offered.versions : v.version]
}
