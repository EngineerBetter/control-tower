variable "zone" {
  type = string
	default = "{{ .Zone }}"
}
variable "tags" {
  type = string
	default = "{{ .Tags }}"
}
variable "project" {
  type = string
	default = "{{ .Project }}"
}
variable "gcpcredentialsjson" {
  type = string
	default = "{{ .GCPCredentialsJSON }}"
}
variable "externalip" {
  type = string
	default = "{{ .ExternalIP }}"
}

variable "deployment" {
  type = string
	default = "{{ .Deployment }}"
}
variable "region" {
  type = string
	default = "{{ .Region }}"
}

variable "db_tier" {
  type = string
	default = "{{ .DBTier }}"
}

variable "db_username" {
  type = string
	default = "{{ .DBUsername }}"
}
variable "db_password" {
  type = string
	default = "{{ .DBPassword }}"
}

variable "db_name" {
  type = string
  default = "{{ .DBName }}"
}

variable "namespace" {
  type = string
  default = "{{ .Namespace }}"
}

variable "source_access_ip" {
  type = string
  default = "{{ .ExternalIP }}"
}

variable "public_cidr" {
  type = string
  default = "{{ .PublicCIDR }}"
}

variable "private_cidr" {
  type = string
  default = "{{ .PrivateCIDR }}"
}

{{if .DNSManagedZoneName }}
variable "dns_managed_zone_name" {
  type = string
  default = "{{ .DNSManagedZoneName }}"
}

variable "dns_record_set_prefix" {
  type = string
  default = "{{ .DNSRecordSetPrefix }}"
}
{{end}}

provider "google" {
    credentials = "{{ .GCPCredentialsJSON }}"
    project = "{{ .Project }}"
    region = var.region
}


terraform {
	backend "gcs" {
		bucket = "{{ .ConfigBucket }}"
	}

    required_providers {
      google = {
        source = "hashicorp/google"
        version = "~> 3.49.0"
      }
    }
}

{{if .DNSManagedZoneName }}
data "google_dns_managed_zone" "dns_zone" {
  name = var.dns_managed_zone_name
}

resource "google_dns_record_set" "dns" {
  managed_zone = data.google_dns_managed_zone.dns_zone.name
  name = "${var.dns_record_set_prefix}${data.google_dns_managed_zone.dns_zone.dns_name}"
  type    = "A"
  ttl     = 60

  rrdatas = [google_compute_address.atc_ip.address]
}
{{end}}

resource "google_compute_router" "nat-router" {
  name    = "${var.deployment}-router"
  region  = var.region
  network = google_compute_network.default.self_link
  bgp {
    asn = 64514
  }
}

resource "google_compute_router_nat" "worker-nat" {
  name                               = "${var.deployment}-worker-nat"
  project                            = var.project
  region                             = var.region
  router                             = google_compute_router.nat-router.name
  nat_ips                            = google_compute_address.nat_ip.*.self_link
  nat_ip_allocate_option             = "MANUAL_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
  subnetwork {
    name                    = google_compute_subnetwork.private.self_link
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
  log_config {
    filter = "TRANSLATIONS_ONLY"
    enable = true
  }
}

resource "google_compute_network" "default" {
  name                    = var.deployment
  project                 = var.project
  auto_create_subnetworks = "false"
}

resource "google_compute_subnetwork" "public" {
  name          = "${var.deployment}-${var.namespace}-public"
  ip_cidr_range = var.public_cidr
  network       = google_compute_network.default.self_link
  project       = var.project
}
resource "google_compute_subnetwork" "private" {
  name          = "${var.deployment}-${var.namespace}-private"
  ip_cidr_range = var.private_cidr
  network       = google_compute_network.default.self_link
  project       = var.project
}

resource "google_compute_firewall" "director" {
  name = "${var.deployment}-director"
  description = "Firewall for external access to BOSH director"
  network     = google_compute_network.default.self_link
  target_tags = ["external"]
  source_ranges = ["${var.source_access_ip}/32", "${google_compute_address.nat_ip.address}/32"]
  allow {
    protocol = "tcp"
    ports = ["6868", "25555", "22"]
  }
}

resource "google_compute_firewall" "atc-http" {
  name = "${var.deployment}-atc-http"
  description = "Firewall for external access to concourse atc"
  network     = google_compute_network.default.self_link
  target_tags = ["web"]
  source_tags = ["web", "worker", "external", "internal"]
  source_ranges = [{{ .AllowIPs }}]
  allow {
    protocol = "tcp"
    ports = ["80"]
  }
}

resource "google_compute_firewall" "atc-https" {
  name = "${var.deployment}-atc-https"
  description = "Firewall for external access to concourse atc"
  network     = google_compute_network.default.self_link
  target_tags = ["web"]
  source_ranges = ["${google_compute_address.nat_ip.address}/32", "${google_compute_address.atc_ip.address}/32", {{ .AllowIPs }}]
  allow {
    protocol = "tcp"
    ports = ["443", "8443"]
  }
}

resource "google_compute_firewall" "from-public" {
  name = "${var.deployment}-public"
  description = "Control-Tower firewall from public VMs"
  network     = google_compute_network.default.self_link
  target_tags = ["web", "external", "internal", "worker"]
  source_ranges = [var.public_cidr]
  allow {
    protocol = "tcp"
    // "25250", "25777", "4222", "22" == BOSH
    // https://github.com/cloudfoundry/bosh-deployment#security-groups
    // 7777 == garden
    // 7788 == baggageclaim
    // 7799 == reaper
    ports = ["25250", "25777", "4222", "22", "7777", "7788", "7799"]
  }
  allow {
    protocol = "udp"
    ports = ["53"]
  }
  allow {
    protocol = "icmp"
  }
}

resource "google_compute_firewall" "from-private" {
  name = "${var.deployment}-private"
  description = "Control-Tower firewall from private VMs"
  network     = google_compute_network.default.self_link
  target_tags = ["web", "external", "internal", "worker"]
  source_ranges = [var.private_cidr]
  allow {
    protocol = "tcp"
    // "25250", "25777", "4222", "22" == BOSH
    // https://github.com/cloudfoundry/bosh-deployment#security-groups
    // 2222 == worker registration
    ports = ["25250", "25777", "4222", "22", "2222"]
  }
{{if .MetricsEnabled}}
  allow {
    protocol = "tcp"
    // Telegraf/InfluxDB
    ports = ["8086"]
  }
{{ end }}
  allow {
    protocol = "udp"
    ports = ["53"]
  }
  allow {
    protocol = "icmp"
  }
}

resource "google_compute_firewall" "atc-services" {
  name = "${var.deployment}-atc-services"
  description = "Firewall for external access to concourse atc"
  network     = google_compute_network.default.self_link
  target_tags = ["web"]
  source_ranges = ["${google_compute_address.nat_ip.address}/32", "${google_compute_address.atc_ip.address}/32", {{ .AllowIPs }}]
  allow {
    protocol = "tcp"
    ports = ["8844"]
  }
{{if .MetricsEnabled}}
  allow {
    protocol = "tcp"
    // Grafana
    ports = ["3000"]
  }
{{ end }}
}

resource "google_compute_firewall" "internal" {
  name        = "${var.deployment}-int"
  description = "BOSH CI Internal Traffic"
  network     = google_compute_network.default.self_link
  source_tags = ["internal"]
  target_tags = ["internal"]

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  allow {
    protocol = "icmp"
  }
}

resource "google_compute_firewall" "sql" {
  name        = "${var.deployment}-sql"
  description = "BOSH CI External Traffic"
  network     = google_compute_network.default.self_link
  direction = "EGRESS"
  allow {
    protocol = "tcp"
    ports    = ["5432"]
  }
  destination_ranges = ["${google_sql_database_instance.director.first_ip_address}/32"]
}

resource "google_service_account" "bosh" {
  account_id   = "${var.deployment}-bosh"
  display_name = "bosh"
}
resource "google_service_account_key" "bosh" {
  service_account_id = google_service_account.bosh.name
  public_key_type = "TYPE_X509_PEM_FILE"
}

resource "google_project_iam_member" "bosh" {
  project = var.project
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.bosh.email}"
}

resource "google_service_account" "self_update" {
  account_id   = "${var.deployment}-su"
  display_name = "self_update"
}
resource "google_service_account_key" "self_update" {
  service_account_id = google_service_account.self_update.name
  public_key_type = "TYPE_X509_PEM_FILE"
}

resource "google_project_iam_member" "self_update" {
  project = var.project
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.self_update.email}"
}

resource "google_compute_address" "atc_ip" {
  name = "${var.deployment}-atc-ip"
}

resource "google_compute_address" "director" {
  name = "${var.deployment}-director-ip"
}

resource "google_compute_address" "nat_ip" {
  name = "${var.deployment}-nat-ip"
}

resource "google_sql_database_instance" "director" {
  name                = var.db_name
  database_version    = "POSTGRES_9_6"
  region              = var.region
  deletion_protection = false

  settings {
    tier = var.db_tier
    user_labels = {
      deployment = var.deployment
    }

    ip_configuration {
      ipv4_enabled = "true"
      authorized_networks {
        name = "atc_conf"
        value = "${google_compute_address.atc_ip.address}/32"
      }

      authorized_networks {
          name = "bosh"
          value = "${google_compute_address.director.address}/32"
      }

      authorized_networks {
          name = "nat"
          value = "${google_compute_address.nat_ip.address}/32"
      }
    }
  }
}

resource "google_sql_database" "director" {
  name      = "udb"
  instance  = google_sql_database_instance.director.name
}

resource "google_sql_user" "director" {
  name     = var.db_username
  instance = google_sql_database_instance.director.name
  password = var.db_password
  deletion_policy = "ABANDON"
}

output "network" {
value = google_compute_network.default.name
}

output "director_firewall_name" {
value = google_compute_firewall.director.name
}

output "private_subnetwork_name" {
value = google_compute_subnetwork.private.name
}

output "public_subnetwork_name" {
value = google_compute_subnetwork.public.name
}

output "private_subnetwork_internal_gw" {
value = google_compute_subnetwork.private.gateway_address
}

output "public_subnetwork_internal_gw" {
value = google_compute_subnetwork.public.gateway_address
}

output "atc_public_ip" {
value = google_compute_address.atc_ip.address
}

output "director_account_creds" {
  value = base64decode(google_service_account_key.bosh.private_key)
  sensitive = true
}

output "self_update_account_creds" {
  value = base64decode(google_service_account_key.self_update.private_key)
  sensitive = true
}

output "director_public_ip" {
  value = google_compute_address.director.address
}

output "bosh_db_address" {
  value = google_sql_database_instance.director.first_ip_address
}

output "db_name" {
  value = google_sql_database_instance.director.name
}

output "nat_gateway_ip" {
  value = google_compute_address.nat_ip.address
}

output "server_ca_cert" {
  value = google_sql_database_instance.director.server_ca_cert.0.cert
}
