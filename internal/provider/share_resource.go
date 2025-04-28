package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/api"
	"github.com/passbolt/go-passbolt/helper"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &shareResource{}
	_ resource.ResourceWithConfigure = &shareResource{}
)

// NewShareResource is a helper function to simplify the provider implementation.
func NewShareResource() resource.Resource {
	return &shareResource{}
}

// shareResource is the resource implementation.
type shareResource struct {
	client *PassboltClient
}

// sharesResourceData create request
type sharesResourceData struct {
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	ShareTargetType  types.String `tfsdk:"share_target_type"`
	ShareTargetValue types.String `tfsdk:"share_target_value"`
	SharePermission  types.String `tfsdk:"share_permission"`
}

// Configure adds the provider configured client to the resource.
func (r *shareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *shareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_share"
}

// Schema defines the schema for the resource.
func (r *shareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Passbolt Share Resource.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the resource to share",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the resource to share, either: resource, folder",
				Computed:    true,
				Default:     stringdefault.StaticString("folder"),
			},
			"share_target_type": schema.StringAttribute{
				Description: "The type of the share target, either: User, Group",
				Required:    true,
			},
			"share_target_value": schema.StringAttribute{
				Description: "The name-value of the share target. Looks up users username/email or groups name",
				Required:    true,
			},
			"share_permission": schema.StringAttribute{
				Description: "The share permission to apply, either: Read: 1, Update: 7, Owner: 15, Delete: -1",
				Required:    true,
			},
		},
	}
}

// Create a new resource.
func (r *shareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan sharesResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.setPermission(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Failed to share resource", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *shareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data sharesResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// read folders permission
	pem, err := r.getPermissionEntry(ctx, data.Name.ValueString(), data.ShareTargetType.ValueString(), data.ShareTargetValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to lookup permission, folder: %s, share-target: %s, share-value: %s", data.Name.ValueString(), data.ShareTargetType.ValueString(), data.ShareTargetValue.ValueString()), err.Error())
		return
	}
	data.SharePermission = types.StringValue(fmt.Sprintf("%d", pem.Type))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *shareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data sharesResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.setPermission(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to update share resource", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *shareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sharesResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	pem, err := r.getPermissionEntry(ctx, data.Name.ValueString(), data.ShareTargetType.ValueString(), data.ShareTargetValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to lookup permission, folder: %s, share-target: %s, share-value: %s", data.Name.ValueString(), data.ShareTargetType.ValueString(), data.ShareTargetValue.ValueString()), err.Error())
		return
	}
	if pem == nil {
		// permission already deleted
		return
	}
	if data.Type.ValueString() == "folder" {
		pem.Delete = true
		err := r.client.Client.ShareFolder(ctx, pem.ACOForeignKey, []api.Permission{*pem})
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("failed to delete permission, folder: %s, share-target: %s, share-value: %s", data.Name.ValueString(), data.ShareTargetType.ValueString(), data.ShareTargetValue.ValueString()), err.Error())
			return
		}
	}
}

func (r *shareResource) getAllFolders(ctx context.Context, search string) ([]api.Folder, error) {
	return r.client.Client.GetFolders(ctx, &api.GetFoldersOptions{FilterSearch: search, ContainPermission: true, ContainPermissions: true, ContainPermissionUserProfile: true, ContainPermissionGroup: true})
}
func (r *shareResource) getAllGroups(ctx context.Context) ([]api.Group, error) {
	return r.client.Client.GetGroups(ctx, &api.GetGroupsOptions{})
}
func (r *shareResource) getAllUsers(ctx context.Context) ([]api.User, error) {
	return r.client.Client.GetUsers(ctx, &api.GetUsersOptions{})
}
func (r *shareResource) getPermissionEntry(ctx context.Context, folderName string, shareTargetType string, shareTargetValue string) (*api.Permission, error) {
	folders, err := r.getAllFolders(ctx, folderName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to lookup folder of: %s, err: %v", folderName, err.Error()))
	}
	groups, err := r.getAllGroups(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to fetch groups, err: %v", err.Error()))
	}
	users, err := r.getAllUsers(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to fetch users, err: %v", err.Error()))
	}
	for _, el := range folders {
		if el.Personal == false {
			for _, pel := range el.Permissions {
				if pel.ACO == "Folder" && pel.ARO == shareTargetType {
					if pel.ARO == "User" {
						for _, uel := range users {
							if uel.Username == shareTargetValue && uel.ID == pel.AROForeignKey {
								return &pel, nil
							}
						}
					} else { // aro=Group
						for _, gel := range groups {
							if gel.Name == shareTargetValue && gel.ID == pel.AROForeignKey {
								return &pel, nil
							}
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func (r *shareResource) setPermission(ctx context.Context, data sharesResourceData) error {
	pemTypeInt := -1
	switch data.SharePermission.ValueString() {
	case "-1":
		pemTypeInt = -1
	case "1":
		pemTypeInt = 1
	case "7":
		pemTypeInt = 7
	case "15":
		pemTypeInt = 15
	default:
		return errors.New(fmt.Sprintf("invalid share permission type, expected one of: -1,1,7,15, got input: %s", data.SharePermission.ValueString()))
	}

	folders, err := r.getAllFolders(ctx, data.Name.ValueString())
	if err != nil {
		return errors.New(fmt.Sprintf("failed to find folder of name: %s, err: %v", data.Name.ValueString(), err.Error()))
	}
	if len(folders) < 1 {
		return errors.New(fmt.Sprintf("failed to find any folder of name: %s", data.Name.ValueString()))
	}
	var folder api.Folder
	for _, el := range folders {
		// TODO: do we want to apply requested permission to all found folders?
		folder = el
		break
	}
	if len(folder.ID) < 0 {
		return errors.New(fmt.Sprintf("failed to find folder of name: %s", data.Name.ValueString()))
	}

	aroID := ""
	if data.ShareTargetType.ValueString() == "Group" {
		groups, err := r.getAllGroups(ctx)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to fetch groups, err: %v", err.Error()))
		}
		for _, el := range groups {
			if el.Name == data.ShareTargetValue.ValueString() {
				aroID = el.ID
				break
			}
		}
	} else {
		users, err := r.getAllUsers(ctx)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to fetch users, err: %v", err.Error()))
		}
		for _, el := range users {
			if el.Username == data.ShareTargetValue.ValueString() {
				aroID = el.ID
				break
			}
		}
	}
	if aroID == "" {
		return errors.New(fmt.Sprintf("failed to find share target, type: %s, value: %s", data.ShareTargetType.ValueString(), data.ShareTargetValue.ValueString()))
	}
	if shareErr := helper.ShareFolder(ctx, r.client.Client, folder.ID, []helper.ShareOperation{
		{
			Type:  pemTypeInt,
			ARO:   data.ShareTargetType.ValueString(),
			AROID: aroID,
		},
	}); shareErr != nil {
		return errors.New(fmt.Sprintf("Failed to share resource, %s, %s, err: %v", folder.ID, aroID, shareErr.Error()))
	}
	return nil
}
