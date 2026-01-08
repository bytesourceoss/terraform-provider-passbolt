data "passbolt_user" "example" {
  username = "dummy@exmpl.com"
}

output "user_out" {
  value = data.passbolt_user.example
}
output "user_out_id" {
  value = data.passbolt_user.example.users[0].id
}
