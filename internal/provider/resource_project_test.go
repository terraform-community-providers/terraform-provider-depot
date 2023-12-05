package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResourceDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfigDefault("todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "depot_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update with null values
			{
				Config: testAccProjectResourceConfigDefault("todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// Update just name
			{
				Config: testAccProjectResourceConfigDefault("nu-todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "nu-todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccProjectResourceConfigNonDefault("nue-todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "nue-todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "depot_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccProjectResourceNonDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfigNonDefault("todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "depot_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update with same values
			{
				Config: testAccProjectResourceConfigNonDefault("todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// Update with null values
			{
				Config: testAccProjectResourceConfigDefault("nue-todo-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("depot_project.test", "id", idRegex()),
					resource.TestCheckResourceAttr("depot_project.test", "name", "nue-todo-app"),
					resource.TestCheckResourceAttr("depot_project.test", "region", "eu-central-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "depot_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProjectResourceConfigDefault(name string) string {
	return fmt.Sprintf(`
resource "depot_project" "test" {
  name = "%s"
  region = "eu-central-1"
}
`, name)
}

func testAccProjectResourceConfigNonDefault(name string) string {
	return fmt.Sprintf(`
resource "depot_project" "test" {
  name = "%s"
  region_id = "eu-central-1"

  cache = {
    size = 100
    expiry = 30
  }
}
`, name)
}
