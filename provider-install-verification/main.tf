terraform {
  required_providers {
    passbolt = {
      source = "opaas-cloud/passbolt"
      version = "1.0.2"
    }
  }
}

provider "passbolt" {
  base_url    = ""
  private_key  = ""
  passphrase = ""
}

resource "passbolt_folder" "example" {
  name = ""
  folder_parent = ""
}

output "example_folders_create" {
  value = passbolt_folder.example
}

resource "passbolt_password" "test" {
  name = ""
  password = ""
  username = ""
  uri = ""
  folder_parent = ""
  share_group = ""
}

