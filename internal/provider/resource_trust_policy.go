package provider

import (
	"context"
	"fmt"
	"strings"

	"buf.build/gen/go/depot/api/connectrpc/go/depot/core/v1/corev1connect"
	corev1 "buf.build/gen/go/depot/api/protocolbuffers/go/depot/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &TrustPolicyResource{}
var _ resource.ResourceWithImportState = &TrustPolicyResource{}

func NewTrustPolicyResource() resource.Resource {
	return &TrustPolicyResource{}
}

type TrustPolicyResource struct {
	client corev1connect.ProjectServiceClient
}

type TrustPolicyGithubResourceModel struct {
	Owner      types.String `tfsdk:"owner"`
	Repository types.String `tfsdk:"repository"`
}

var githubAttrTypes = map[string]attr.Type{
	"owner":      types.StringType,
	"repository": types.StringType,
}

type TrustPolicyBuildkiteResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	Pipeline     types.String `tfsdk:"pipeline"`
}

var buildkiteAttrTypes = map[string]attr.Type{
	"organization": types.StringType,
	"pipeline":     types.StringType,
}

type TrustPolicyCircleciResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	Project      types.String `tfsdk:"project"`
}

var circleciAttrTypes = map[string]attr.Type{
	"organization": types.StringType,
	"project":      types.StringType,
}

type TrustPolicyResourceModel struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	Github    types.Object `tfsdk:"github"`
	Buildkite types.Object `tfsdk:"buildkite"`
	Circleci  types.Object `tfsdk:"circleci"`
}

func (r *TrustPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_policy"
}

func (r *TrustPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Depot trust policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the trust policy.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the project for the trust policy.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.UTF8LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"github": schema.SingleNestedAttribute{
				MarkdownDescription: "GitHub provider settings for the trust policy.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"owner": schema.StringAttribute{
						MarkdownDescription: "GitHub owner name.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
					},
					"repository": schema.StringAttribute{
						MarkdownDescription: "GitHub repository name.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(path.MatchRoot("buildkite"), path.MatchRoot("circleci")),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"buildkite": schema.SingleNestedAttribute{
				MarkdownDescription: "Buildkite provider settings for the trust policy.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"organization": schema.StringAttribute{
						MarkdownDescription: "Buildkite organization slug.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
					},
					"pipeline": schema.StringAttribute{
						MarkdownDescription: "Buildkite pipeline slug.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"circleci": schema.SingleNestedAttribute{
				MarkdownDescription: "CircleCI provider settings for the trust policy.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"organization": schema.StringAttribute{
						MarkdownDescription: "CircleCI organization uuid.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
					},
					"project": schema.StringAttribute{
						MarkdownDescription: "CircleCI project uuid.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *TrustPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TrustPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TrustPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := corev1.AddTrustPolicyRequest{
		ProjectId: data.ProjectId.ValueString(),
	}

	if !data.Github.IsNull() {
		var githubData *TrustPolicyGithubResourceModel

		resp.Diagnostics.Append(data.Github.As(ctx, &githubData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.Provider = &corev1.AddTrustPolicyRequest_Github{
			Github: &corev1.TrustPolicy_GitHub{
				RepositoryOwner: githubData.Owner.ValueString(),
				Repository:      githubData.Repository.ValueString(),
			},
		}
	} else if !data.Buildkite.IsNull() {
		var buildkiteData *TrustPolicyBuildkiteResourceModel

		resp.Diagnostics.Append(data.Buildkite.As(ctx, &buildkiteData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.Provider = &corev1.AddTrustPolicyRequest_Buildkite{
			Buildkite: &corev1.TrustPolicy_Buildkite{
				OrganizationSlug: buildkiteData.Organization.ValueString(),
				PipelineSlug:     buildkiteData.Pipeline.ValueString(),
			},
		}
	} else if !data.Circleci.IsNull() {
		var circleciData *TrustPolicyCircleciResourceModel

		resp.Diagnostics.Append(data.Circleci.As(ctx, &circleciData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.Provider = &corev1.AddTrustPolicyRequest_Circleci{
			Circleci: &corev1.TrustPolicy_CircleCI{
				OrganizationUuid: circleciData.Organization.ValueString(),
				ProjectUuid:      circleciData.Project.ValueString(),
			},
		}
	} else {
		resp.Diagnostics.AddError("Invalid Plan", "Trust policy must have exactly one provider.")
		return
	}

	response, err := r.client.AddTrustPolicy(ctx, &connect.Request[corev1.AddTrustPolicyRequest]{
		Msg: &input,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create trust policy, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created a trust policy")

	data.Id = types.StringValue(response.Msg.TrustPolicy.TrustPolicyId)
	data.ProjectId = types.StringValue(data.ProjectId.ValueString())

	if response.Msg.TrustPolicy.GetGithub() != nil {
		data.Github = types.ObjectValueMust(
			githubAttrTypes,
			map[string]attr.Value{
				"owner":      types.StringValue(response.Msg.TrustPolicy.GetGithub().RepositoryOwner),
				"repository": types.StringValue(response.Msg.TrustPolicy.GetGithub().Repository),
			},
		)
	} else if response.Msg.TrustPolicy.GetBuildkite() != nil {
		data.Buildkite = types.ObjectValueMust(
			buildkiteAttrTypes,
			map[string]attr.Value{
				"organization": types.StringValue(response.Msg.TrustPolicy.GetBuildkite().OrganizationSlug),
				"pipeline":     types.StringValue(response.Msg.TrustPolicy.GetBuildkite().PipelineSlug),
			},
		)
	} else if response.Msg.TrustPolicy.GetCircleci() != nil {
		data.Circleci = types.ObjectValueMust(
			circleciAttrTypes,
			map[string]attr.Value{
				"organization": types.StringValue(response.Msg.TrustPolicy.GetCircleci().OrganizationUuid),
				"project":      types.StringValue(response.Msg.TrustPolicy.GetCircleci().ProjectUuid),
			},
		)
	} else {
		resp.Diagnostics.AddError("Invalid Response", "Trust policy must have exactly one provider.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TrustPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TrustPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.ListTrustPolicies(ctx, &connect.Request[corev1.ListTrustPoliciesRequest]{
		Msg: &corev1.ListTrustPoliciesRequest{
			ProjectId: data.ProjectId.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list trust policies, got error: %s", err))
		return
	}

	trustPolicy, err := findTrustPolicy(ctx, response.Msg.TrustPolicies, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find trust policy, got error: %s", err))
		return
	}

	data.Id = types.StringValue(trustPolicy.TrustPolicyId)
	data.ProjectId = types.StringValue(data.ProjectId.ValueString())

	if trustPolicy.GetGithub() != nil {
		data.Github = types.ObjectValueMust(
			githubAttrTypes,
			map[string]attr.Value{
				"owner":      types.StringValue(trustPolicy.GetGithub().RepositoryOwner),
				"repository": types.StringValue(trustPolicy.GetGithub().Repository),
			},
		)
	} else if trustPolicy.GetBuildkite() != nil {
		data.Buildkite = types.ObjectValueMust(
			buildkiteAttrTypes,
			map[string]attr.Value{
				"organization": types.StringValue(trustPolicy.GetBuildkite().OrganizationSlug),
				"pipeline":     types.StringValue(trustPolicy.GetBuildkite().PipelineSlug),
			},
		)
	} else if trustPolicy.GetCircleci() != nil {
		data.Circleci = types.ObjectValueMust(
			circleciAttrTypes,
			map[string]attr.Value{
				"organization": types.StringValue(trustPolicy.GetCircleci().OrganizationUuid),
				"project":      types.StringValue(trustPolicy.GetCircleci().ProjectUuid),
			},
		)
	} else {
		resp.Diagnostics.AddError("Invalid Response", "Trust policy must have exactly one provider.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TrustPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state *TrustPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TrustPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TrustPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RemoveTrustPolicy(ctx, &connect.Request[corev1.RemoveTrustPolicyRequest]{
		Msg: &corev1.RemoveTrustPolicyRequest{
			ProjectId:     data.ProjectId.ValueString(),
			TrustPolicyId: data.Id.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete trust policy, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a trust policy")
}

func (r *TrustPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: project_id:trust_policy_id. Got: %q", req.ID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
}

func findTrustPolicy(ctx context.Context, policies []*corev1.TrustPolicy, id string) (*corev1.TrustPolicy, error) {
	for _, policy := range policies {
		if policy.TrustPolicyId == id {
			return policy, nil
		}
	}

	return nil, fmt.Errorf("trust policy doesn't exist")
}
