go_package(
    name="packages",
)

go_binary(
    name="terraform-provider-passbolt",
)

go_mod(
    name="mod",
)

file(
    name="terraformrc",
    source="passbolt.tfrc"
)

run_shell_command(
    name="terraform-test",
    command="scripts/terraform_test.sh {chroot}",
    execution_dependencies=[":terraformrc"],
)
