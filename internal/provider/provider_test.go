// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	p := New("test")()
	resp := provider.MetadataResponse{}

	p.Metadata(context.TODO(), provider.MetadataRequest{}, &resp)

	// We use the provider type in other files, so it should be set correctly.
	assert.Equal(t, resp.TypeName, "passbolt")
}

func TestSchema(t *testing.T) {
	p := New("test")()
	resp := provider.SchemaResponse{}

	p.Schema(context.TODO(), provider.SchemaRequest{}, &resp)

	// Let's make sure these are set to sensitive
	assert.True(t, resp.Schema.Attributes["private_key"].IsSensitive())
	assert.True(t, resp.Schema.Attributes["passphrase"].IsSensitive())
}

func TestConfigure(t *testing.T) {
	p := New("test")()
	resp := provider.ConfigureResponse{}

	os.Setenv("PASSBOLT_URL", "https://test.example.com")
	os.Setenv("PASSBOLT_KEY", "--- TEST KEY ---")
	os.Setenv("PASSBOLT_PASS", "TestKeyPassword")

	p.Configure(context.TODO(), provider.ConfigureRequest{}, &resp)
}
