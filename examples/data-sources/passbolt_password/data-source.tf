# Copyright (c) HashiCorp, Inc.

# Gets all folders
data "passbolt_password" "my_secret" {
  id = "00000000-1111-2222-3333-444444444444"
}

output "my_secret_password" {
  # The value will still be hidden, as it's classified as a `sensative` string.
  value = data.passbolt_password.my_secret.password
}
