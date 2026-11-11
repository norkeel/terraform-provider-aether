data "norkeel_k8s_versions" "offered" {}

resource "norkeel_cluster" "prod" {
  name   = "prod"
  region = "eu-fi"
  # latest is null while the platform's version window is unavailable; pin a
  # minor instead (e.g. "1.37") if that's not acceptable.
  k8s_version = data.norkeel_k8s_versions.offered.latest
  # Defaults to true: the public API endpoint is billed as one public IP (€4/month).
  public_endpoint = true
}

# Fixed pool: a flat monthly fee for desired_nodes nodes.
resource "norkeel_node_pool" "workers" {
  cluster_id    = norkeel_cluster.prod.id
  name          = "workers"
  vcpus         = 4
  memory_gb     = 16
  disk_gb       = 80
  desired_nodes = 3
}

# Autoscaling pool: a vCPU/memory budget with max > min makes it autoscaling (billed hourly per node).
resource "norkeel_node_pool" "burst" {
  cluster_id    = norkeel_cluster.prod.id
  name          = "burst"
  vcpus         = 4
  memory_gb     = 16
  disk_gb       = 80
  min_vcpus     = 4
  max_vcpus     = 32
  min_memory_gb = 16
  max_memory_gb = 128
}

data "norkeel_kubeconfig" "prod" {
  cluster_id = norkeel_cluster.prod.id
}

output "kubeconfig" {
  value     = data.norkeel_kubeconfig.prod.content
  sensitive = true
}
