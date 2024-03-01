package provider

import (
	"context"
	"fmt"

	"buf.build/gen/go/depot/api/connectrpc/go/depot/core/v1/corev1connect"
	corev1 "buf.build/gen/go/depot/api/protocolbuffers/go/depot/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

const sizeGB = 1024 * 1024 * 1024

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client corev1connect.ProjectServiceClient
}

type ProjectResourceCacheModel struct {
	Size   types.Int64 `tfsdk:"size"`
	Expiry types.Int64 `tfsdk:"expiry"`
}

var cacheAttrTypes = map[string]attr.Type{
	"size":   types.Int64Type,
	"expiry": types.Int64Type,
}

type ProjectResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Region         types.String `tfsdk:"region"`
	Cache          types.Object `tfsdk:"cache"`
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Depot project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the project.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the organization.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.UTF8LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the project.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.UTF8LengthAtLeast(1),
					stringvalidator.UTF8LengthAtMost(64),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region of the project.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.UTF8LengthAtLeast(1),
				},
			},
			"cache": schema.SingleNestedAttribute{
				MarkdownDescription: "Cache policy of the project.",
				Optional:            true,
				Computed:            true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					cacheAttrTypes,
					map[string]attr.Value{
						"size":   types.Int64Value(50),
						"expiry": types.Int64Value(14),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						MarkdownDescription: "Number of bytes to keep in the cache in GB. **Default** `50`.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(50),
					},
					"expiry": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep the cache for. **Default** `14`.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(14),
					},
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*corev1connect.ProjectServiceClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = *client
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProjectResourceModel
	var cacheData *ProjectResourceCacheModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.Cache.As(ctx, &cacheData, basetypes.ObjectAsOptions{})...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := corev1.CreateProjectRequest{
		Name:     data.Name.ValueString(),
		RegionId: data.Region.ValueString(),
		CachePolicy: &corev1.CachePolicy{
			KeepBytes: cacheData.Size.ValueInt64() * sizeGB,
			KeepDays:  int32(cacheData.Expiry.ValueInt64()),
		},
	}

	if !data.OrganizationId.IsNull() {
		input.OrganizationId = data.OrganizationId.ValueStringPointer()
	}

	response, err := r.client.CreateProject(ctx, &connect.Request[corev1.CreateProjectRequest]{
		Msg: &input,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create project, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created a project")

	data.Id = types.StringValue(response.Msg.Project.ProjectId)
	data.OrganizationId = types.StringValue(response.Msg.Project.OrganizationId)
	data.Name = types.StringValue(response.Msg.Project.Name)
	data.Region = types.StringValue(response.Msg.Project.RegionId)

	data.Cache = types.ObjectValueMust(
		cacheAttrTypes,
		map[string]attr.Value{
			"size":   types.Int64Value(response.Msg.Project.CachePolicy.KeepBytes / sizeGB),
			"expiry": types.Int64Value(int64(response.Msg.Project.CachePolicy.KeepDays)),
		},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.GetProject(ctx, &connect.Request[corev1.GetProjectRequest]{
		Msg: &corev1.GetProjectRequest{
			ProjectId: data.Id.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project, got error: %s", err))
		return
	}

	data.Id = types.StringValue(response.Msg.Project.ProjectId)
	data.OrganizationId = types.StringValue(response.Msg.Project.OrganizationId)
	data.Name = types.StringValue(response.Msg.Project.Name)
	data.Region = types.StringValue(response.Msg.Project.RegionId)

	data.Cache = types.ObjectValueMust(
		cacheAttrTypes,
		map[string]attr.Value{
			"size":   types.Int64Value(response.Msg.Project.CachePolicy.KeepBytes / sizeGB),
			"expiry": types.Int64Value(int64(response.Msg.Project.CachePolicy.KeepDays)),
		},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProjectResourceModel
	var cacheData *ProjectResourceCacheModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.Cache.As(ctx, &cacheData, basetypes.ObjectAsOptions{})...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := corev1.UpdateProjectRequest{
		ProjectId: data.Id.ValueString(),
		Name:      data.Name.ValueStringPointer(),
		RegionId:  data.Region.ValueStringPointer(),
		CachePolicy: &corev1.CachePolicy{
			KeepBytes: cacheData.Size.ValueInt64() * sizeGB,
			KeepDays:  int32(cacheData.Expiry.ValueInt64()),
		},
	}

	response, err := r.client.UpdateProject(ctx, &connect.Request[corev1.UpdateProjectRequest]{
		Msg: &input,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update project, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated a project")

	data.Id = types.StringValue(response.Msg.Project.ProjectId)
	data.OrganizationId = types.StringValue(response.Msg.Project.OrganizationId)
	data.Name = types.StringValue(response.Msg.Project.Name)
	data.Region = types.StringValue(response.Msg.Project.RegionId)

	data.Cache = types.ObjectValueMust(
		cacheAttrTypes,
		map[string]attr.Value{
			"size":   types.Int64Value(response.Msg.Project.CachePolicy.KeepBytes / sizeGB),
			"expiry": types.Int64Value(int64(response.Msg.Project.CachePolicy.KeepDays)),
		},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteProject(ctx, &connect.Request[corev1.DeleteProjectRequest]{
		Msg: &corev1.DeleteProjectRequest{
			ProjectId: data.Id.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete project, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a project")
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
