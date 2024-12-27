# Copyright (c) HashiCorp, Inc.

# Basic configuration
provider "passbolt" {
  base_url    = "https://example.passbolt.com"    # PASSBOLT_URL
  private_key = "<YOUR PASSBOLT PGP PRIVATE KEY>" # PASSBOLT_KEY
  passphrase  = "<YOUR PASSBOLT PASSPHRASE>"      # PASSBOLT_PASS
}
