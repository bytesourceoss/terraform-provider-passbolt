terraform {
  required_providers {
    passbolt = {
      source = "hashicorp.com/edu/passbolt"
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
}

