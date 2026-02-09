package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/groteck/terraform-provider-pangolin/internal/client"
)

var _ resource.Resource = &resourceResource{}
var _ resource.ResourceWithImportState = &resourceResource{}

func NewResourceResource() resource.Resource {
	return &resourceResource{}
}

type resourceResource struct {
	client *client.Client
}

type resourceResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	OrgID     types.String `tfsdk:"org_id"`
	Name      types.String `tfsdk:"name"`
	Protocol  types.String `tfsdk:"protocol"`
	Http      types.Bool   `tfsdk:"http"`
	Subdomain types.String `tfsdk:"subdomain"`
	DomainID  types.String `tfsdk:"domain_id"`
}

func (r *resourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *resourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an app-style resource (HTTP/TCP/UDP).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the resource.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the resource.",
			},
			"protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The protocol of the resource (tcp or udp).",
				Validators: []validator.String{
					stringvalidator.OneOf("tcp", "udp"),
				},
			},
			"http": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the resource is an HTTP resource.",
			},
			"subdomain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The subdomain for the resource.",
			},
			"domain_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the domain.",
			},
		},
	}
}

func (r *resourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *resourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := &client.Resource{
		Name:      data.Name.ValueString(),
		Protocol:  data.Protocol.ValueString(),
		Http:      data.Http.ValueBool(),
		Subdomain: data.Subdomain.ValueString(),
		DomainID:  data.DomainID.ValueString(),
	}

	created, err := r.client.CreateResource(data.OrgID.ValueString(), res)
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(created.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetResource(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading resource", err.Error())
		return
	}

	data.Name = types.StringValue(res.Name)
	data.Protocol = types.StringValue(res.Protocol)
	data.Http = types.BoolValue(res.Http)
	data.Subdomain = types.StringValue(res.Subdomain)
	data.DomainID = types.StringValue(res.DomainID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state resourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := &client.Resource{
		Name:      data.Name.ValueString(),
		Protocol:  data.Protocol.ValueString(),
		Http:      data.Http.ValueBool(),
		Subdomain: data.Subdomain.ValueString(),
		DomainID:  data.DomainID.ValueString(),
	}

	_, err := r.client.UpdateResource(int(state.ID.ValueInt64()), res)
	if err != nil {
		resp.Diagnostics.AddError("Error updating resource", err.Error())
		return
	}

	data.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting resource", err.Error())
		return
	}
}

func (r *resourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org_id/id
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id/id. Got: %q", req.ID),
		)
		return
	}

	resID, err := strconv.ParseInt(idParts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected id to be an integer. Got: %q", idParts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resID)...)
}
