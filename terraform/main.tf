# Copyright (c) HashiCorp, Inc.

# Used for testing

terraform {
  required_providers {
    passbolt = {
      source = "riebecj/passbolt"
    }
  }
}

resource "passbolt_password" "minimum" {
  name = "Test Minimum Password Resource"
  username = "test-user-min"
  password = "test-password-min"
  uri = "https://min.test.internal"
}

resource "passbolt_password" "all" {
  name = "Test All Password Resource"
  description = "Test Description"
  password = "test-password-all"
  username = "test-user-all"
  uri = "https://all.test.internal"
  folder_parent = "TestFolder"
  share_group = "TestShareGroup"
}
