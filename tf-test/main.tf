terraform {
  required_providers {
    passbolt = {
      source  = "bytesourceoss/passbolt"
      version = "1.0.0"
    }
  }
}

data "passbolt_user" "admin_user" {
  username = "admin@example.com"
}

output "user_out" {
  value = data.passbolt_user.admin_user
}
resource "passbolt_group" "group_open" {
  name = "group-open"
  group_users = [
    {
      user_id  = data.passbolt_user.admin_user.users[0].id
      is_admin = true
    },
  ]
}
resource "passbolt_group" "group_open_foo" {
  name = "group-open-foo"
  group_users = [
    {
      user_id  = data.passbolt_user.admin_user.users[0].id
      is_admin = true
    },
  ]
}
resource "passbolt_group" "group_open_bar" {
  name = "group-open-bar"
  group_users = [
    {
      user_id  = data.passbolt_user.admin_user.users[0].id
      is_admin = true
    },
  ]
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

data "passbolt_folder" "folder_shared" {
  name = "folder-shared"
}

resource "passbolt_folder" "folder_shared_foo" {
  name = "folder-shared-foo"
  #folder_parent_id = data.passbolt_folder.folder_shared.id
  folder_parent_id = passbolt_folder.folder_shared.id
}
resource "passbolt_folder" "folder_shared_bar" {
  name             = "folder-shared-bar"
  folder_parent_id = passbolt_folder.folder_shared.id
}


data "passbolt_folder" "folder_shared_bar" {
  name = "folder-shared-bar"
}
data "passbolt_group" "group_open_bar" {
  name = "group-open-bar"
}

resource "passbolt_share" "share_folder_with_group_open" {
  name               = "folder-shared"
  share_target_type  = "Group"
  share_target_value = "group-open"
  share_permission   = "1"
}

resource "passbolt_share" "share_folder_shared_foo_with_group_open_foo" {
  share_source_id   = passbolt_folder.folder_shared_foo.id
  share_target_type = "Group"
  share_target_id   = passbolt_group.group_open_foo.id
  share_permission  = "1"
}

resource "passbolt_share" "share_folder_shared_bar_with_group_open_bar" {
  share_source_id   = passbolt_folder.folder_shared_bar.id
  share_target_type = "Group"
  share_target_id   = passbolt_group.group_open_bar.id
  share_permission  = "1"
}



data "passbolt_share" "all" {}

output "share" {
  # `value` will be a list of all available share
  value = data.passbolt_share.all
}

output "share1" {
  value = passbolt_share.share_folder_with_group_open
}
output "share2" {
  value = passbolt_share.share_folder_shared_foo_with_group_open_foo
}
output "share3" {
  value = passbolt_share.share_folder_shared_bar_with_group_open_bar
}
