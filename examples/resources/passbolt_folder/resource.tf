# Basic Passbolt Folder
resource "passbolt_folder" "basic" {
  name = "My Folder"
}

# Full Passbolt Folder Configuration
resource "passbolt_folder" "full" {
  name          = "My Folder"
  personal      = true
  folder_parent = "Parent Folder"
}
