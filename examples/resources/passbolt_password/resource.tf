# Copyright (c) HashiCorp, Inc.

# Basic Configuration
resource "random_password" "basic" {
  length = 16
}

resource "passbolt_password" "basic" {
  name     = "Basic Password Example"
  username = "myUser"
  password = random_password.basic.result
  uri      = "https://example.com"
}

# Full Password Cofiguration
resource "passbolt_password" "full" {
  name          = "Full Password Example"
  description   = "A Description for the secret."
  username      = "myUser"
  password      = random_password.basic.result
  uri           = "https://example.com"
  share_group   = "SomeShareGroup"
  folder_parent = "Parent Folder"
}
