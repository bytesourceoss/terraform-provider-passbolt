data "passbolt_user" "example" {
  username = "dummy@exmpl.com"
}

output "user_out" {
  value = data.passbolt_user.example
}
