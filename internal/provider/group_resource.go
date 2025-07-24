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
	_ resource.Resource              = &groupResource{}
	_ resource.ResourceWithConfigure = &groupResource{}
)

// NewGroupResource is a helper function to simplify the provider implementation.
func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// groupResource is the resource implementation.
type groupResource struct {
	client *PassboltClient
}

// created, modified
type groupModel struct {
	ID         types.String      `tfsdk:"id"`
	Name       types.String      `tfsdk:"name"`
	GroupUsers []groupMembership `tfsdk:"group_users"`
}

type groupMembership struct {
	UserID  types.String `tfsdk:"user_id"`
	IsAdmin types.Bool   `tfsdk:"is_admin"`
	Delete  types.Bool   `tfsdk:"delete"`
}

// Configure adds the provider configured client to the resource.
func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Passbolt User Resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The group id.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The group name.",
				Required:    true,
			},
			"group_users": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Description: "The id of the user to add.",
							Required:    true,
						},
						"is_admin": schema.BoolAttribute{
							Description: "Whether to make the user admin of the group.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"delete": schema.BoolAttribute{
							Description: "Whether the user should be deleted from the group",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
		},
	}
}

// Create a new resource.
func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan groupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	members := make([]api.GroupMembership, 0)
	for _, member := range plan.GroupUsers {
		elem := api.GroupMembership{
			UserID:  member.UserID.ValueString(),
			IsAdmin: member.IsAdmin.ValueBool(),
			Delete:  member.Delete.ValueBool(),
		}

		members = append(members, elem)
	}

	// Generate API request body from plan
	var group = api.Group{
		Name:       plan.Name.ValueString(),
		GroupUsers: members,
	}

	cGroup, errCreate := r.client.Client.CreateGroup(r.client.Context, group)
	if errCreate != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to create group of name: %s", group.Name),
			errCreate.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(cGroup.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state groupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := &api.GetGroupsOptions{
		ContainGroupsUsers: true,
	}
	groups, err := r.client.Client.GetGroups(r.client.Context, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to get groups"),
			err.Error(),
		)
		return
	}
	for _, group := range groups {
		if state.ID.ValueString() == group.ID {
			state.Name = types.StringValue(group.Name)
			members := make([]groupMembership, 0)
			for _, member := range group.GroupUsers {
				elem := groupMembership{
					UserID:  types.StringValue(member.UserID),
					IsAdmin: types.BoolValue(member.IsAdmin),
					Delete:  types.BoolValue(member.Delete),
				}

				members = append(members, elem)
			}
			state.GroupUsers = members
			break
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan groupModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	members := make([]api.GroupMembership, 0)
	for _, member := range plan.GroupUsers {
		elem := api.GroupMembership{
			UserID:  member.UserID.ValueString(),
			IsAdmin: member.IsAdmin.ValueBool(),
			Delete:  member.Delete.ValueBool(),
		}

		members = append(members, elem)
	}

	var update = api.GroupUpdate{
		Name:         plan.Name.ValueString(),
		GroupChanges: members,
	}

	var state groupModel
	req.State.Get(ctx, &state)
	cGroup, err := r.client.Client.UpdateGroup(r.client.Context, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to update group of name: %s", state.Name.ValueString()),
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(cGroup.ID)
	state.Name = types.StringValue(cGroup.Name)
	cMembers := make([]groupMembership, 0)
	for _, member := range cGroup.GroupUsers {
		elem := groupMembership{
			UserID:  types.StringValue(member.UserID),
			IsAdmin: types.BoolValue(member.IsAdmin),
			Delete:  types.BoolValue(member.Delete),
		}

		cMembers = append(cMembers, elem)
	}
	state.GroupUsers = cMembers

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from plan
	var state groupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Client.DeleteGroup(r.client.Context, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to delete group with ID: %s", state.ID.ValueString()),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
