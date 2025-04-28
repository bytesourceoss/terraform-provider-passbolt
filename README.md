# Passbolt Terraform Provider

Uses [go-passbolt](https://github.com/passbolt/go-passbolt) as the Passbolt client to your Passbolt instance. 

## local development

create file `~/.terraformrc` or `~/.tofurc` & fill with this config to make tf or tofu lookup provider locally.
Go path can be looked up by: `echo $GOPATH`

```
provider_installation {

  dev_overrides {
    "bytesourceoss/passbolt" = "<<INSERT_GO_PATH_HERE>>/bin/"
  }

  direct {}
}
```

create provider & install
`make all`

run terraform/tofu as usual
