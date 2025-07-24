package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource is the resource implementation.
type userResource struct {
	client *PassboltClient
}

// created, modified
type usersModel struct {
	ID        types.String `tfsdk:"id"`
	Role      types.String `tfsdk:"role"`
	UserName  types.String `tfsdk:"username"`
	FirstName types.String `tfsdk:"firstname"`
	LastName  types.String `tfsdk:"lastname"`
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Passbolt User Resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The user id.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The user role.",
				Computed:    true,
				Optional:    true,
				//Default:     stringdefault.StaticString(""),
			},
			"username": schema.StringAttribute{
				Description: "The user name. This needs to be a valid email.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"firstname": schema.StringAttribute{
				Description: "The first name of the user.",
				Required:    true,
			},
			"lastname": schema.StringAttribute{
				Description: "The last name of the user.",
				Required:    true,
			},
		},
	}
}

// Create a new resource.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan usersModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var user = api.User{
		Username: plan.UserName.ValueString(),
		Profile: &api.Profile{
			FirstName: plan.FirstName.ValueString(),
			LastName:  plan.LastName.ValueString(),
		},
		RoleID: plan.Role.ValueString(),
	}

	cUser, errCreate := r.client.Client.CreateUser(r.client.Context, user)
	if errCreate != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to create user of name: %s", user.Username),
			errCreate.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(cUser.ID)
	plan.Role = types.StringValue(cUser.Role.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state usersModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.Client.GetUser(r.client.Context, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Cannot get user: %s", user.Username),
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(user.ID)
	state.Role = types.StringValue(user.Role.ID)
	state.UserName = types.StringValue(user.Username)
	state.FirstName = types.StringValue(user.Profile.FirstName)
	state.LastName = types.StringValue(user.Profile.LastName)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan usersModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var user = api.User{
		Username: plan.UserName.ValueString(),
		Profile: &api.Profile{
			FirstName: plan.FirstName.ValueString(),
			LastName:  plan.LastName.ValueString(),
		},
		RoleID: plan.Role.ValueString(),
	}

	var state usersModel
	req.State.Get(ctx, &state)
	cUser, err := r.client.Client.UpdateUser(r.client.Context, state.ID.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to update user of name: %s", user.Username),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(cUser.ID)
	plan.Role = types.StringValue(cUser.Role.ID)
	plan.UserName = types.StringValue(cUser.Username)
	plan.FirstName = types.StringValue(cUser.Profile.FirstName)
	plan.LastName = types.StringValue(cUser.Profile.LastName)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from plan
	var state usersModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Client.DeleteUser(r.client.Context, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to delete user with ID: %s", state.ID.ValueString()),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
