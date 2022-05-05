variable "cluster_name" {
  type = string
  default = "monitoring-cluster"
}

variable "project" {
  type = string
  default = "idyllic-silicon-343409"
}

variable "location" {
  type = string
  default = "europe-west3-a"
}

variable "initial_node_count" {
  type = number
  default = 1
}

variable "machine_type" {
  type = string
  default = "e2-standard-8"
}