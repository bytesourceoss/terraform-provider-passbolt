#!/bin/bash

main() {
    pantsSandbox=$1
    gopath=$(go env | grep GOPATH | cut -d"=" -f2 | xargs)
    sed -i -e "s|<GOPATH>|${gopath}|" "${pantsSandbox}/passbolt.tfrc"
    go install .
    cd tools
    go generate ./...
}

main "$@"
