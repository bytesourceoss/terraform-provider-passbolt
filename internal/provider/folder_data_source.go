package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &folderDataSource{}
	_ datasource.DataSourceWithConfigure = &folderDataSource{}
)

// NewFolderDataSource is a helper function to simplify the provider implementation.
func NewFolderDataSource() datasource.DataSource {
	return &folderDataSource{}
}

// coffeesDataSource is the data source implementation.
type folderDataSource struct {
	client *PassboltClient
}

type folderDataSourceModel foldersModel

// Configure adds the provider configured client to the data source.
func (d *folderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *folderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

// Schema defines the schema for the data source.
func (d *folderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
			"modified": schema.StringAttribute{
				Computed: true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"modified_by": schema.StringAttribute{
				Computed: true,
			},
			"folder_parent_id": schema.StringAttribute{
				Computed: true,
			},
			"personal": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *folderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state folderDataSourceModel

	folder, err := d.client.Client.GetFolders(d.client.Context, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read folder", "",
		)
		return
	}

	var reqModel folderDataSourceModel

	diags := req.Config.Get(ctx, &reqModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	filterValue := reqModel.Name.ValueString()
	if filterValue == "" {
		resp.Diagnostics.AddError(
			"folder name missing", "",
		)
		return
	}

	var folderList = make([]folderDataSourceModel, 0)

	// Map response body to model
	for _, el := range folder {
		folderEl := folderDataSourceModel{
			ID:             types.StringValue(el.ID),
			Name:           types.StringValue(el.Name),
			Created:        types.StringValue(el.Created.String()),
			Modified:       types.StringValue(el.Modified.String()),
			CreatedBy:      types.StringValue(el.CreatedBy),
			ModifiedBy:     types.StringValue(el.ModifiedBy),
			FolderParentId: types.StringValue(el.FolderParentID),
			Personal:       types.BoolValue(el.Personal),
		}
		if filterValue == folderEl.Name.ValueString() {
			folderList = append(folderList, folderEl)
		}
	}

	if len(folderList) > 1 {
		resp.Diagnostics.AddError(
			"multiple folders found, abort", "",
		)
		return
	}
	if len(folderList) > 0 {
		state = folderList[0]
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
