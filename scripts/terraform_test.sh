#!/bin/bash

main() {
    pantsSandbox=$1
    gopath=$(go env | grep GOPATH | cut -d"=" -f2 | xargs)
    sed -i -e "s|<GOPATH>|${gopath}|" "${pantsSandbox}/passbolt.tfrc"
    go install .
    TF_CLI_CONFIG_FILE="${pantsSandbox}/passbolt.tfrc" PASSBOLT_URL="test-url" PASSBOLT_KEY="test-key" PASSBOLT_PASS="test-pass" terraform -chdir=terraform test
}

main "$@"
