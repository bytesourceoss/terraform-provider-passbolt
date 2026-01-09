# Share Passbolt folder with user (readonly)
resource "passbolt_share" "share-folder-with-user" {
  name               = "folder-name"
  share_target_type  = "User"
  share_target_value = "test@user.com"
  share_permission   = "1"
}

# Share Passbolt folder with group (update)
resource "passbolt_share" "share-folder-with-group" {
  name               = "shared-folder-name"
  share_target_type  = "Group"
  share_target_value = "shared-group"
  share_permission   = "7"
}

# Un-Share Passbolt folder from group (delete share)
resource "passbolt_share" "share-folder-with-group" {
  name               = "non-shared-folder-name"
  share_target_type  = "Group"
  share_target_value = "shared-group"
  share_permission   = "-1"
}

data "passbolt_folder" "folder_shared_foobar" {
  name = "folder-shared-foobar"
}
data "passbolt_group" "group_open_foobar" {
  name = "group-open-foobar"
}
data "passbolt_user" "john_doe" {
  username = "john_doe@example.com"
}


resource "passbolt_share" "share-folder-with-user" {
  share_source_id   = data.passbolt_folder.folder_shared_foobar.id
  share_target_type = "Group"
  share_target_id   = data.passbolt_group.group_open_foobar.id
  share_permission  = "1"
}
resource "passbolt_share" "share-folder-with-user" {
  share_source_id   = data.passbolt_folder.folder_shared_foobar.id
  share_target_type = "User"
  share_target_id   = data.passbolt_user.john_doe.users[0].id
  share_permission  = "1"
}
