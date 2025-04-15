package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &shareDataSource{}
	_ datasource.DataSourceWithConfigure = &shareDataSource{}
)

// NewShareDataSource is a helper function to simplify the provider implementation.
func NewShareDataSource() datasource.DataSource {
	return &shareDataSource{}
}

// shareDataSource is the data source implementation.
type shareDataSource struct {
	client *PassboltClient
}

type shareDataSourceModel struct {
	Shares []shareModel `tfsdk:"shares"`
}

type shareModel struct {
	ID         types.String `tfsdk:"id"`
	RoleID     types.String `tfsdk:"role_id"`
	Name       types.String `tfsdk:"name"`
	Username   types.String `tfsdk:"username"`
	Active     types.Bool   `tfsdk:"active"`
	Deleted    types.Bool   `tfsdk:"deleted"`
	Created    types.String `tfsdk:"created"`
	Modified   types.String `tfsdk:"modified"`
	CreatedBy  types.String `tfsdk:"created_by"`
	ModifiedBy types.String `tfsdk:"modified_by"`
}

// Configure adds the provider configured client to the data source.
func (d *shareDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *shareDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_share"
}

// Schema defines the schema for the data source.
func (d *shareDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"share": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"role_id": schema.StringAttribute{
							Required: false,
						},
						"name": schema.StringAttribute{
							Required: false,
						},
						"username": schema.StringAttribute{
							Required: false,
						},
						"active": schema.BoolAttribute{
							Required: false,
						},
						"deleted": schema.BoolAttribute{
							Required: false,
						},
						"created": schema.StringAttribute{
							Required: true,
						},
						"modified": schema.StringAttribute{
							Required: true,
						},
						"created_by": schema.StringAttribute{
							Required: true,
						},
						"modified_by": schema.StringAttribute{
							Required: false,
						},
						"folder_parent_id": schema.StringAttribute{
							Required: false,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *shareDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state shareDataSourceModel
	var data shareModel
	diag := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diag...)

	shares, err := d.client.Client.SearchAROs(d.client.Context, api.SearchAROsOptions{FilterSearch: data.ID.String()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read shares", "",
		)
		return
	}

	// Map response body to model
	for _, share := range shares {
		shareState := shareModel{
			ID:         types.StringValue(share.User.ID),
			RoleID:     types.StringValue(share.RoleID),
			Name:       types.StringValue(share.Name),
			Username:   types.StringValue(share.Username),
			Active:     types.BoolValue(share.Active),
			Deleted:    types.BoolValue(share.User.Deleted),
			Created:    types.StringValue(share.User.Created.String()),
			Modified:   types.StringValue(share.User.Modified.String()),
			CreatedBy:  types.StringValue(share.CreatedBy),
			ModifiedBy: types.StringValue(share.ModifiedBy),
		}

		state.Shares = append(state.Shares, shareState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
