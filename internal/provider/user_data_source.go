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
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource is the datasource implementation.
type userDataSource struct {
	client *PassboltClient
}

type usersDataSourceModel struct {
	UserName types.String     `tfsdk:"username"`
	Users    []usersReadModel `tfsdk:"users"`
}

// created, modified
type usersReadModel struct {
	ID           types.String `tfsdk:"id"`
	UserName     types.String `tfsdk:"username"`
	Profile      profileModel `tfsdk:"profile"`
	Created      types.String `tfsdk:"created"`
	Modified     types.String `tfsdk:"modified"`
	Active       types.Bool   `tfsdk:"active"`
	Deleted      types.Bool   `tfsdk:"deleted"`
	RoleID       types.String `tfsdk:"role_id"`
	Description  types.String `tfsdk:"description"`
	LastLoggedIn types.String `tfsdk:"last_logged_in"`
}

type profileModel struct {
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
}

// Configure adds the provider configured client to the datasource.
func (r *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Metadata returns the datasource type name.
func (r *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the datasource.
func (r *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Passbolt User DataSource.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "The user name. This needs to be a valid email. User will be replaced upon username change.",
				Optional:    true,
			},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"username": schema.StringAttribute{
							Required: true,
						},
						"created": schema.StringAttribute{
							Required: true,
						},
						"modified": schema.StringAttribute{
							Required: true,
						},
						"active": schema.BoolAttribute{
							Required: true,
						},
						"deleted": schema.BoolAttribute{
							Required: true,
						},
						"role_id": schema.StringAttribute{
							Required: true,
						},
						"description": schema.StringAttribute{
							Required: true,
						},
						"last_logged_in": schema.StringAttribute{
							Required: true,
						},
						"profile": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"first_name": schema.StringAttribute{
									Required: true,
								},
								"last_name": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from state
	var reqModel, state usersDataSourceModel
	diags := req.Config.Get(ctx, &reqModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := r.client.Client.GetUsers(r.client.Context, &api.GetUsersOptions{
		FilterSearch: reqModel.UserName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Cannot get user: %s", reqModel.UserName.ValueString()),
			err.Error(),
		)
		return
	}

	for _, user := range users {
		userState := usersReadModel{
			ID:       types.StringValue(user.ID),
			UserName: types.StringValue(user.Username),
			Profile: profileModel{
				FirstName: types.StringValue(user.Profile.FirstName),
				LastName:  types.StringValue(user.Profile.LastName),
			},
			Created:      types.StringValue(user.Created.String()),
			Modified:     types.StringValue(user.Modified.String()),
			Active:       types.BoolValue(user.Active),
			Deleted:      types.BoolValue(user.Deleted),
			RoleID:       types.StringValue(user.RoleID),
			Description:  types.StringValue(user.Description),
			LastLoggedIn: types.StringValue(user.LastLoggedIn),
		}
		state.Users = append(state.Users, userState)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
