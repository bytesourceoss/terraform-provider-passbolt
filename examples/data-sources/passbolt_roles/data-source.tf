resource "passbolt_user" "dummy" {
  username  = "dummy@test.com"
  firstname = "Johnny"
  lastname  = "Test"
  role      = local.role_admin_id
}

data "passbolt_roles" "all" {}

output "out" {
  value = data.passbolt_roles.all.roles
}

locals {
  role_admin_id = one([for item in data.passbolt_roles.all.roles : item if item.name == "admin"]).id
  role_user_id  = one([for item in data.passbolt_roles.all.roles : item if item.name == "user"]).id
}
