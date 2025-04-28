# Gets all folders
data "passbolt_folders" "all" {}

# Get all folders matching regex
data "passbolt_folders" "folder_named_example" {
  name = "example"
}

output "folders" {
  # `value` will be a list of all available folders
  value = data.passbolt_folders.all
}
