package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/helper"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &passwordResource{}
	_ resource.ResourceWithConfigure = &passwordResource{}
)

// NewPasswordResource is a helper function to simplify the provider implementation.
func NewPasswordResource() resource.Resource {
	return &passwordResource{}
}

// folderResource is the resource implementation.
type passwordResource struct {
	client *PassboltClient
}

type passwordModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Username       types.String `tfsdk:"username"`
	Uri            types.String `tfsdk:"uri"`
	ShareGroup     types.String `tfsdk:"share_group"`
	FolderParent   types.String `tfsdk:"folder_parent"`
	FolderParentId types.String `tfsdk:"folder_parent_id"`
	Password       types.String `tfsdk:"password"`
}

// Configure adds the provider configured client to the resource.
func (r *passwordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PassboltClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *passwordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

// Schema defines the schema for the resource.
func (r *passwordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defines a Passbolt Secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Resource ID of the secret.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the secret.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the secret",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username of the secret.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "The URI of the secret.",
				Required:    true,
			},
			"share_group": schema.StringAttribute{
				Description: "The Group Name to share the secret with.",
				Optional:    true,
			},
			"folder_parent": schema.StringAttribute{
				Description: "The parent folder in which to place the secret.",
				Optional:    true,
			},
			"folder_parent_id": schema.StringAttribute{
				Description: "The ID of the parent folder, if `folder_parent` is specified.",
				Optional:    true,
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "The secret password, stored as a sensative string in state.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

// Create a new resource.
func (r *passwordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan passwordModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folders, errFolder := r.client.Client.GetFolders(ctx, nil)
	if errFolder != nil {
		resp.Diagnostics.AddError("Cannot get folders", errFolder.Error())
		return
	}

	if !plan.FolderParent.IsUnknown() && !plan.FolderParent.IsNull() {
		for _, folder := range folders {
			if folder.Name == plan.FolderParent.ValueString() {
				plan.FolderParentId = types.StringValue(folder.ID)
			}
		}
	}

	resourceId, err := helper.CreateResource(
		ctx,
		r.client.Client,
		plan.FolderParentId.ValueString(),
		plan.Name.ValueString(),
		plan.Username.ValueString(),
		plan.Uri.ValueString(),
		plan.Password.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Cannot create resource", err.Error())
		return
	}

	var groupId string
	if !plan.ShareGroup.IsUnknown() && !plan.FolderParent.IsNull() {
		groups, _ := r.client.Client.GetGroups(ctx, nil)

		for _, group := range groups {
			if group.Name == plan.ShareGroup.ValueString() {
				groupId = group.ID
			}
		}

		if groupId != "" {
			var shares = []helper.ShareOperation{
				{
					Type:  7,
					ARO:   "Group",
					AROID: groupId,
				},
			}

			shareErr := helper.ShareResource(ctx, r.client.Client, resourceId, shares)

			if shareErr != nil {
				resp.Diagnostics.AddError("Cannot share resource", shareErr.Error())
			}
		}
	}

	plan.ID = types.StringValue(resourceId)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *passwordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state passwordModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	folderParentID, name, username, uri, password, description, err := helper.GetResource(r.client.Context, r.client.Client, state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if folderParentID != "" {
		_, folderName, err := helper.GetFolder(r.client.Context, r.client.Client, folderParentID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get folder name for "+folderParentID, err.Error(),
			)
			return
		}
		state.FolderParent = types.StringValue(folderName)
	}

	if description != "" {
		state.Description = types.StringValue(description)
	}
	state.Name = types.StringValue(name)
	state.Username = types.StringValue(username)
	state.Password = types.StringValue(password)
	state.Uri = types.StringValue(uri)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *passwordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state passwordModel
	var plan passwordModel
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	planDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Resource
	err := helper.UpdateResource(
		r.client.Context,
		r.client.Client,
		state.ID.ValueString(),
		plan.Name.ValueString(),
		plan.Username.ValueString(),
		plan.Uri.ValueString(),
		plan.Password.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating resource "+state.ID.ValueString(), err.Error(),
		)
		return
	}
	if plan.FolderParent.ValueString() != state.FolderParent.ValueString() {
		folders, err := r.client.Client.GetFolders(r.client.Context, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read folders", err.Error(),
			)
			return
		}
		for _, folder := range folders {
			if folder.Name == plan.FolderParent.ValueString() {
				err = helper.MoveResource(r.client.Context, r.client.Client, state.ID.ValueString(), folder.ID)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error moving resource "+state.ID.ValueString()+" to folder "+folder.ID, err.Error(),
					)
					return
				}
			}
		}
	}
	if plan.ShareGroup.ValueString() != state.ShareGroup.ValueString() {
		permissedUsers := make([]string, 0)
		permissedGroups := make([]string, 0)
		groups, err := r.client.Client.GetGroups(r.client.Context, nil)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Groups", err.Error())
			return
		}
		for _, group := range groups {
			if group.Name == plan.ShareGroup.ValueString() {
				permissedGroups = append(permissedGroups, group.ID)
			}
		}
		if len(permissedGroups) < 1 {
			resp.Diagnostics.AddError(
				"Unable to find Grou ID for "+plan.ShareGroup.ValueString(), "",
			)
			return
		}
		err = helper.ShareResourceWithUsersAndGroups(r.client.Context, r.client.Client, state.ID.ValueString(), permissedUsers, permissedGroups, 1)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to share "+state.ID.ValueString()+" with "+plan.ShareGroup.ValueString(), err.Error(),
			)
			return
		}
	}

	// Read updated data
	folderParentID, name, username, uri, password, description, err := helper.GetResource(r.client.Context, r.client.Client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read resource "+state.ID.ValueString(), err.Error(),
		)
		return
	}

	if description != "" {
		state.Description = types.StringValue(description)
	} else {
		state.Description = types.StringNull()
	}
	state.Name = types.StringValue(name)
	state.Username = types.StringValue(username)
	state.Password = types.StringValue(password)
	state.Uri = types.StringValue(uri)
	state.FolderParent = types.StringValue(plan.FolderParent.ValueString())
	state.FolderParentId = types.StringValue(folderParentID)
	state.ShareGroup = plan.ShareGroup

	setStateDiags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(setStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *passwordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state passwordModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.client.Client.DeleteResource(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting password", err.Error(),
		)
		return
	}
}
