package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTrustPolicyResourceDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTrustPolicyResourceConfigDefault(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("depot_trust_policy.test", "id"),
					resource.TestCheckResourceAttr("depot_trust_policy.test", "project_id", "7f76t7vghb"),
					resource.TestCheckResourceAttr("depot_trust_policy.test", "github.owner", "terraform-community-providers"),
					resource.TestCheckResourceAttr("depot_trust_policy.test", "github.repository", "terraform-provider-depot"),
					resource.TestCheckNoResourceAttr("depot_trust_policy.test", "buildkite"),
					resource.TestCheckNoResourceAttr("depot_trust_policy.test", "circleci"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "depot_trust_policy.test",
				ImportState:       true,
				ImportStateIdFunc: trustPolicyImportIdFunc,
				ImportStateVerify: true,
			},
			// Update with same values
			{
				Config: testAccTrustPolicyResourceConfigDefault(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("depot_trust_policy.test", "id"),
					resource.TestCheckResourceAttr("depot_trust_policy.test", "project_id", "7f76t7vghb"),
					resource.TestCheckResourceAttr("depot_trust_policy.test", "github.owner", "terraform-community-providers"),
					resource.TestCheckResourceAttr("depot_trust_policy.test", "github.repository", "terraform-provider-depot"),
					resource.TestCheckNoResourceAttr("depot_trust_policy.test", "buildkite"),
					resource.TestCheckNoResourceAttr("depot_trust_policy.test", "circleci"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "depot_trust_policy.test",
				ImportState:       true,
				ImportStateIdFunc: trustPolicyImportIdFunc,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTrustPolicyResourceConfigDefault() string {
	return `
resource "depot_trust_policy" "test" {
  project_id = "7f76t7vghb"

  github = {
    owner      = "terraform-community-providers"
    repository = "terraform-provider-depot"
  }
}
`
}

func trustPolicyImportIdFunc(state *terraform.State) (string, error) {
	rawState, ok := state.RootModule().Resources["depot_trust_policy.test"]

	if !ok {
		return "", fmt.Errorf("Resource Not found")
	}

	return fmt.Sprintf("%s:%s", rawState.Primary.Attributes["project_id"], rawState.Primary.Attributes["id"]), nil
}
