package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &folderResource{}
	_ resource.ResourceWithConfigure = &folderResource{}
)

// NewFolderResource is a helper function to simplify the provider implementation.
func NewFolderResource() resource.Resource {
	return &folderResource{}
}

// folderResource is the resource implementation.
type folderResource struct {
	client *PassboltClient
}

// created, modified
type foldersModelCreate struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Personal types.Bool   `tfsdk:"personal"`
}

// Configure adds the provider configured client to the resource.
func (r *folderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Metadata returns the resource type name.
func (r *folderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

// Schema defines the schema for the resource.
func (r *folderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Passbolt Folder Resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The folder Resource ID.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The folder name",
				Required:    true,
			},
			"personal": schema.BoolAttribute{
				Description: "If the folder is a personal folder.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// Create a new resource.
func (r *folderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan foldersModelCreate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folders, err := r.client.Client.GetFolders(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Cannot get folders", "")
		return
	}

	/* TODO: parent folder id on creation does not work, check if we need a move into parent folder feature
	var parentFolderID string
	if !plan.FolderParent.IsUnknown() && !plan.FolderParent.IsNull() {
		for _, folder := range folders {
			if folder.Name == plan.FolderParent.ValueString() {
				parentFolderID = folder.ID
			}
		}
	}
	*/

	// Generate API request body from plan
	var folder = api.Folder{
		Name:     plan.Name.ValueString(),
		Personal: plan.Personal.ValueBool(),
	}

	for _, el := range folders {
		if el.Name == folder.Name && el.Personal == folder.Personal {
			// folder already created
			return
		}
	}

	cFolder, errCreate := r.client.Client.CreateFolder(r.client.Context, folder)
	if errCreate != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to create folder of name: %s", folder.Name),
			errCreate.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(cFolder.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *folderResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *folderResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *folderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
