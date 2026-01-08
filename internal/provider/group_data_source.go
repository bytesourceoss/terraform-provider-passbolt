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
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

// NewGroupDataSource is a helper function to simplify the provider implementation.
func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

// coffeesDataSource is the data source implementation.
type groupDataSource struct {
	client *PassboltClient
}

type groupDataSourceModel struct {
	ID         types.String      `tfsdk:"id"`
	Name       types.String      `tfsdk:"name"`
	GroupUsers []groupMembership `tfsdk:"group_users"`
	Deleted    types.Bool        `tfsdk:"deleted"`
	Created    types.String      `tfsdk:"created"`
	Modified   types.String      `tfsdk:"modified"`
	CreatedBy  types.String      `tfsdk:"created_by"`
	ModifiedBy types.String      `tfsdk:"modified_by"`
	UserCount  types.Int64       `tfsdk:"user_count"`
}

// Configure adds the provider configured client to the data source.
func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the data source.
func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"deleted": schema.BoolAttribute{
				Computed: true,
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
			"user_count": schema.Int64Attribute{
				Computed: true,
			},
			"group_users": schema.ListNestedAttribute{
				Required: false,
				Computed: true,
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
						},
						"delete": schema.BoolAttribute{
							Description: "Whether the user should be deleted from the group",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupDataSourceModel

	group, err := d.client.Client.GetGroups(d.client.Context, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read group", "",
		)
		return
	}

	var reqModel groupDataSourceModel

	diags := req.Config.Get(ctx, &reqModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	filterValue := reqModel.Name.ValueString()
	if filterValue == "" {
		resp.Diagnostics.AddError(
			"group name missing", "",
		)
		return
	}

	var groupList = make([]groupDataSourceModel, 0)

	// Map response body to model
	for _, el := range group {
		members := make([]groupMembership, 0)
		for _, member := range el.GroupUsers {
			members = append(members, groupMembership{
				UserID:  types.StringValue(member.UserID),
				IsAdmin: types.BoolValue(member.IsAdmin),
				Delete:  types.BoolValue(member.Delete),
			})
		}
		state.GroupUsers = members
		groupEl := groupDataSourceModel{
			ID:         types.StringValue(el.ID),
			Name:       types.StringValue(el.Name),
			GroupUsers: members,
			Deleted:    types.BoolValue(el.Deleted),
			Created:    types.StringValue(el.Created.String()),
			Modified:   types.StringValue(el.Modified.String()),
			CreatedBy:  types.StringValue(el.CreatedBy),
			ModifiedBy: types.StringValue(el.ModifiedBy),
			UserCount:  types.Int64Value(int64(len(el.Users))),
		}
		if filterValue == groupEl.Name.ValueString() {
			groupList = append(groupList, groupEl)
		}
	}

	if len(groupList) > 1 {
		resp.Diagnostics.AddError(
			"multiple groups found, abort", "",
		)
		return
	}
	if len(groupList) > 0 {
		state = groupList[0]
	}
	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
