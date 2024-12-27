# Copyright (c) HashiCorp, Inc.

# Gets all folders
data "passbolt_folders" "all" {}

output "folders" {
  # `value` will be a list of all available folders
  value = data.passbolt_folders.all
}