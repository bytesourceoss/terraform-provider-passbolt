# Basic Passbolt Folder
resource "passbolt_folder" "basic" {
  name = "My Folder"
}

data "passbolt_folder" "parent_folder" {
  name = "parent_folder"
}

# Full Passbolt Folder Configuration
resource "passbolt_folder" "full" {
  name             = "My Folder"
  folder_parent_id = data.passbolt_folder.parent_folder.id
}
