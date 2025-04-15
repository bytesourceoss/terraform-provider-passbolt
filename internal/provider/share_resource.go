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

// sharesModelCreate create request
type sharesModelCreate struct {
	ID            types.String `tfsdk:"id"`
	Aro           types.String `tfsdk:"aro"`
	AroForeignKey types.String `tfsdk:"aro_foreign_key"`
	Type          types.Number `tfsdk:"type"`
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
	resp.TypeName = req.ProviderTypeName + "_folder"
}

// Schema defines the schema for the resource.
func (r *shareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Passbolt Share Resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The folder Resource ID.",
				Computed:    true,
			},
			"aro": schema.StringAttribute{
				Description: "ARO: User or Group",
				Required:    false,
			},
			"aro_foreign_key": schema.StringAttribute{
				Description: "ARO id, User-id, Group-id",
				Required:    false,
			},
			"type": schema.NumberAttribute{
				Description: "permission type: Read: 1, Update: 7, Owner: 15",
				Required:    true,
			},
		},
	}
}

// Create a new resource.
func (r *shareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan sharesModelCreate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	typ, _ := plan.Type.ValueBigFloat().Int(nil)
	if shareErr := helper.ShareFolder(ctx, r.client.Client, plan.ID.String(), []helper.ShareOperation{
		{
			Type:  int(typ.Int64()),
			ARO:   plan.Aro.String(),
			AROID: plan.AroForeignKey.String(),
		},
	}); shareErr != nil {
		resp.Diagnostics.AddError("Cannot share resource", shareErr.Error())
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
func (r *shareResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *shareResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *shareResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
