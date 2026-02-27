package provider

import (
	"context"
	"fmt"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &resourceOrganization{}
var _ resource.ResourceWithImportState = &resourceOrganization{}

func NewOrganizationResource() resource.Resource {
	return &resourceOrganization{}
}

type resourceOrganization struct {
	client *client.Client
}

type resourceOrganizationModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Subnet        types.String `tfsdk:"subnet"`
	UtilitySubnet types.String `tfsdk:"utility_subnet"`
}

func (r *resourceOrganization) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *resourceOrganization) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages organizations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the organization.",
			},
			"subnet": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"utility_subnet": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The utility subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceOrganization) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Organization Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *resourceOrganizationModel) ValueOrganization() client.Organization {
	res := client.Organization{
		Name:          r.Name.ValueString(),
		Subnet:        r.Subnet.ValueStringPointer(),
		UtilitySubnet: r.UtilitySubnet.ValueStringPointer(),
	}
	return res
}

func (r *resourceOrganization) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceOrganizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := data.ValueOrganization()
	org.ID = data.ID.ValueString()
	_, err := r.client.CreateOrganization(org)
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceOrganization) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceOrganizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetOrganization(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	data.ID = types.StringValue(res.ID)
	data.Name = types.StringValue(res.Name)
	data.Subnet = types.StringValue(*res.Subnet)
	data.UtilitySubnet = types.StringValue(*res.UtilitySubnet)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceOrganization) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state resourceOrganizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateOrganization(
		state.ID.ValueString(),
		data.ValueOrganization(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization", err.Error())
		return
	}

	data.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceOrganization) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceOrganizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrganization(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting organization", err.Error())
		return
	}
}

func (r *resourceOrganization) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org_id
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), req.ID)...)
}
