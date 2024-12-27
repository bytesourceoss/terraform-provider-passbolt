// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/helper"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &passwordDataSource{}
	_ datasource.DataSourceWithConfigure = &passwordDataSource{}
)

// NewPasswordDataSource is a helper function to simplify the provider implementation.
func NewPasswordDataSource() datasource.DataSource {
	return &passwordDataSource{}
}

// passwordDataSource is the data source implementation.
type passwordDataSource struct {
	client *PassboltClient
}

type passwordDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Username       types.String `tfsdk:"username"`
	Uri            types.String `tfsdk:"uri"`
	FolderParentID types.String `tfsdk:"folder_parent_id"`
	Password       types.String `tfsdk:"password"`
}

// Configure adds the provider configured client to the data source.
func (d *passwordDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PassboltClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *passboltClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Metadata returns the data source type name.
func (d *passwordDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

// Schema defines the schema for the data source.
func (d *passwordDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets a Passbolt secret for the provided Resource ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Passbolt Resource ID of the secret (can be seen at the end of the URL of the secret in the web UI).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the secret.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the secret. If not defined, it returns an empty string.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username of the secret.",
				Computed:    true,
			},
			"uri": schema.StringAttribute{
				Description: "The URI of the secret.",
				Computed:    true,
			},
			"folder_parent_id": schema.StringAttribute{
				Description: "The ID of the parent folder, if any. Otherwise it's an empty string.",
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "The decrypted password of the secret.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *passwordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data passwordDataSourceModel
	diag := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diag...)

	folderParentID, name, username, uri, password, description, err := helper.GetResource(d.client.Context, d.client.Client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read resource "+data.ID.ValueString(), err.Error(),
		)
		return
	}

	data.Name = types.StringValue(name)
	data.Description = types.StringValue(description)
	data.Uri = types.StringValue(uri)
	data.Username = types.StringValue(username)
	data.FolderParentID = types.StringValue(folderParentID)
	data.Password = types.StringValue(password)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
