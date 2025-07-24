resource "passbolt_group" "grp" {
  name = "devops-team"

  group_users = [
    {
      user_id  = "5f8642a0-f3e3-403b-b666-8cda965fbad6"
      is_admin = true
    },
  ]
}
