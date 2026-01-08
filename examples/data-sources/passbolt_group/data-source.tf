# Get group matching name exactly or error
data "passbolt_group" "group_name_exact_match_example" {
  name = "exact_match_example"
}

output "group" {
  # `value` will be the group
  value = data.passbolt_group.group_name_exact_match_example
}
