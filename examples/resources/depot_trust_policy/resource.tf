resource "depot_trust_policy" "example" {
  project_id = "wkgrl762gp"

  github {
    owner      = "example"
    repository = "example"
  }
}
