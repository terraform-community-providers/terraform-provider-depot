---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "depot_trust_policy Resource - terraform-provider-depot"
subcategory: ""
description: |-
  Depot trust policy.
---

# depot_trust_policy (Resource)

Depot trust policy.

## Example Usage

```terraform
resource "depot_trust_policy" "example" {
  project_id = "wkgrl762gp"

  github {
    owner      = "example"
    repository = "example"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) Identifier of the project for the trust policy.

### Optional

- `buildkite` (Attributes) Buildkite provider settings for the trust policy. (see [below for nested schema](#nestedatt--buildkite))
- `circleci` (Attributes) CircleCI provider settings for the trust policy. (see [below for nested schema](#nestedatt--circleci))
- `github` (Attributes) GitHub provider settings for the trust policy. (see [below for nested schema](#nestedatt--github))

### Read-Only

- `id` (String) Identifier of the trust policy.

<a id="nestedatt--buildkite"></a>
### Nested Schema for `buildkite`

Required:

- `organization` (String) Buildkite organization slug.
- `pipeline` (String) Buildkite pipeline slug.


<a id="nestedatt--circleci"></a>
### Nested Schema for `circleci`

Required:

- `organization` (String) CircleCI organization uuid.
- `project` (String) CircleCI project uuid.


<a id="nestedatt--github"></a>
### Nested Schema for `github`

Required:

- `owner` (String) GitHub owner name.
- `repository` (String) GitHub repository name.

## Import

Import is supported using the following syntax:

```shell
terraform import depot_trust_policy.example wkgrl762gp:nj38erubf2
```
