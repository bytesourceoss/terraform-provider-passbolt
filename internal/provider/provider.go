package provider

import (
	"context"
	"os"
	"terraform-provider-passbolt/tools"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &passboltProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &passboltProvider{
			version: version,
		}
	}
}

// hashicupsProvider is the provider implementation.
type passboltProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type hashicupsProviderModel struct {
	URL  types.String `tfsdk:"base_url"`
	KEY  types.String `tfsdk:"private_key"`
	PASS types.String `tfsdk:"passphrase"`
}

// Metadata returns the provider type name.
func (p *passboltProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "passbolt"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *passboltProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required: true,
			},
			"private_key": schema.StringAttribute{
				Required: true,
			},
			"passphrase": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *passboltProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config hashicupsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown URL",
			"",
		)
	}

	if config.KEY.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key"),
			"Unknown private key",
			"",
		)
	}

	if config.PASS.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown password",
			"",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	url := os.Getenv("PASSBOLT_URL")
	key := os.Getenv("PASSBOLT_KEY")
	pass := os.Getenv("PASSBOLT_PASS")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.KEY.IsNull() {
		key = config.KEY.ValueString()
	}

	if !config.PASS.IsNull() {
		pass = config.PASS.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing url",
			"",
		)
	}

	if key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Missing private key",
			"",
		)
	}

	if pass == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing password",
			"",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := api.NewClient(nil, "", url, key, pass)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to passbolt",
			"Client Error: "+err.Error(),
		)
		return
	}

	passboltClient := tools.PassboltClient{
		Client:     client,
		Url:        url,
		Context:    context.TODO(),
		Password:   pass,
		PrivateKey: key,
	}

	// Make the client available during DataSource and Resource
	// type Configure methods.

	tools.Login(&passboltClient)

	resp.DataSourceData = &passboltClient
	resp.ResourceData = &passboltClient
}

// DataSources defines the data sources implemented in the provider.
func (p *passboltProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFoldersDataSource,
		NewPasswordDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *passboltProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFolderResource,
		NewPasswordResource,
	}
}
