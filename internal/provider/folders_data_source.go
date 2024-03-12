package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-passbolt/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &foldersDataSource{}
	_ datasource.DataSourceWithConfigure = &foldersDataSource{}
)

// NewFoldersDataSource is a helper function to simplify the provider implementation.
func NewFoldersDataSource() datasource.DataSource {
	return &foldersDataSource{}
}

// coffeesDataSource is the data source implementation.
type foldersDataSource struct {
	client *tools.PassboltClient
}

type foldersDataSourceModel struct {
	Folders []foldersModel `tfsdk:"folders"`
}

type foldersModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Created        types.String `tfsdk:"created"`
	Modified       types.String `tfsdk:"modified"`
	CreatedBy      types.String `tfsdk:"created_by"`
	ModifiedBy     types.String `tfsdk:"modified_by"`
	FolderParentId types.String `tfsdk:"folder_parent_id"`
	Personal       types.Bool   `tfsdk:"personal"`
}

// Configure adds the provider configured client to the data source.
func (d *foldersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tools.PassboltClient)
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
func (d *foldersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folders"
}

// Schema defines the schema for the data source.
func (d *foldersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"folders": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
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
							Required: true,
						},
						"folder_parent_id": schema.StringAttribute{
							Required: true,
						},
						"personal": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *foldersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state foldersDataSourceModel

	folders, err := d.client.Client.GetFolders(d.client.Context, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read folders", "",
		)
		return
	}

	// Map response body to model
	for _, folder := range folders {
		folderState := foldersModel{
			ID:             types.StringValue(folder.ID),
			Name:           types.StringValue(folder.Name),
			Created:        types.StringValue(folder.Created.String()),
			Modified:       types.StringValue(folder.Modified.String()),
			CreatedBy:      types.StringValue(folder.CreatedBy),
			ModifiedBy:     types.StringValue(folder.ModifiedBy),
			FolderParentId: types.StringValue(folder.FolderParentID),
			Personal:       types.BoolValue(folder.Personal),
		}
		state.Folders = append(state.Folders, folderState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
