# Get folder matching name exactly or error
data "passbolt_folder" "folder_name_exact_match_example" {
  name = "exact_match_example"
}

output "folder" {
  # `value` will be the folder
  value = data.passbolt_folder.folder_name_exact_match_example
}
