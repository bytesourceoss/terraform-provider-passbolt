terraform {
  required_providers {
    passbolt = {
      source  = "bytesourceoss/passbolt"
      version = "1.0.0"
    }
  }
}

variable "passbolt_base_url" {
  type = string
}
variable "passbolt_private_key" {
  type = string
}
variable "passbolt_passphrase" {
  type = string
}

provider "passbolt" {
  base_url    = var.passbolt_base_url
  private_key = var.passbolt_private_key
  passphrase  = var.passbolt_passphrase
}

resource "passbolt_folder" "folder_private" {
  name = "folder-private"
}
resource "passbolt_folder" "folder_shared" {
  name = "folder-shared"
}

resource "passbolt_share" "share-folder-with-group-open" {
  name               = "folder-shared"
  share_target_type  = "Group"
  share_target_value = "group-open"
  share_permission   = "1"
}
