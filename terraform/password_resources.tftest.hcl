# Copyright (c) HashiCorp, Inc.

mock_provider "passbolt" {}

run "minimum_password_resource" {
  command = plan

  assert {
    condition = passbolt_password.minimum.name == "Test Minimum Password Resource"
    error_message = "Incorrect Password Resource Name"
  }
  assert {
    condition = passbolt_password.minimum.username == "test-user-min"
    error_message = "Incorrect Username"
  }
  assert {
    condition = passbolt_password.minimum.password == "test-password-min"
    error_message = "Incorrect Password"
  }
  assert {
    condition = passbolt_password.minimum.uri == "https://min.test.internal"
    error_message = "Incorrect URI"
  }
}

run "all_password_resource" {
  command = plan

  assert {
    condition = passbolt_password.all.name == "Test All Password Resource"
    error_message = "Incorrect Password Resource Name"
  }
  assert {
    condition = passbolt_password.all.description == "Test Description"
    error_message = "Incorrect Password Resource Name"
  }
  assert {
    condition = passbolt_password.all.username == "test-user-all"
    error_message = "Incorrect Username"
  }
  assert {
    condition = passbolt_password.all.password == "test-password-all"
    error_message = "Incorrect Password"
  }
  assert {
    condition = passbolt_password.all.uri == "https://all.test.internal"
    error_message = "Incorrect URI"
  }
  assert {
    condition = passbolt_password.all.folder_parent == "TestFolder"
    error_message = "Incorrect Folder"
  }
  assert {
    condition = passbolt_password.all.share_group == "TestShareGroup"
    error_message = "Incorrect Share Group"
  }
}
