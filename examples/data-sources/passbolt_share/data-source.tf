# Gets all share
data "passbolt_share" "all" {}

output "share" {
  # `value` will be a list of all available share
  value = data.passbolt_share.all
}
