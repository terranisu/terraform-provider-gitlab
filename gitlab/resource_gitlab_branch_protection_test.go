package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabBranchProtection_basic(t *testing.T) {

	var pb gitlab.ProtectedBranch
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabBranchProtectionDestroy,
		Steps: []resource.TestStep{
			// Create a project and Branch Protection with default options
			{
				Config: testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.BranchProtect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Update the Branch Protection
			{
				Config: testAccGitlabBranchProtectionUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.BranchProtect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.MasterPermissions],
						MergeAccessLevel: accessLevel[gitlab.MasterPermissions],
					}),
				),
			},
			// Update the Branch Protection to get back to initial settings
			{
				Config: testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.BranchProtect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabBranchProtectionExists(n string, pb *gitlab.ProtectedBranch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}
		project, branch, err := projectAndBranchFromID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error in Splitting Project and Branch Ids")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		pbs, _, err := conn.ProtectedBranches.ListProtectedBranches(project, nil)
		if err != nil {
			return err
		}
		for _, gotpb := range pbs {
			if gotpb.Name == branch {
				*pb = *gotpb
				return nil
			}
		}
		return fmt.Errorf("Protected Branch does not exist")
	}
}

type testAccGitlabBranchProtectionExpectedAttributes struct {
	Name             string
	PushAccessLevel  string
	MergeAccessLevel string
}

func testAccCheckGitlabBranchProtectionAttributes(pb *gitlab.ProtectedBranch, want *testAccGitlabBranchProtectionExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pb.Name != want.Name {
			return fmt.Errorf("got name %q; want %q", pb.Name, want.Name)
		}

		if pb.PushAccessLevels[0].AccessLevel != accessLevelID[want.PushAccessLevel] {
			return fmt.Errorf("got Push access levels %q; want %q", pb.PushAccessLevels[0].AccessLevel, accessLevelID[want.PushAccessLevel])
		}

		if pb.MergeAccessLevels[0].AccessLevel != accessLevelID[want.MergeAccessLevel] {
			return fmt.Errorf("got Merge access levels %q; want %q", pb.MergeAccessLevels[0].AccessLevel, accessLevelID[want.MergeAccessLevel])
		}

		return nil
	}
}

func testAccCheckGitlabBranchProtectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)
	var project string
	var branch string
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "gitlab_project" {
			project = rs.Primary.ID
		} else if rs.Type == "gitlab_branch_protection" {
			branch = rs.Primary.ID
		}
	}

	pb, response, err := conn.ProtectedBranches.GetProtectedBranch(project, branch)
	if err == nil {
		if pb != nil {
			return fmt.Errorf("project branch protection %s still exists", branch)
		}
	}
	if response.StatusCode != 404 {
		return err
	}
	return nil
}

func testAccGitlabBranchProtectionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_branch_protection" "BranchProtect" {
  project = "${gitlab_project.foo.id}"
  branch = "BranchProtect-%d"
  push_access_level = "developer"
  merge_access_level = "developer"
}
	`, rInt, rInt)
}

func testAccGitlabBranchProtectionUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_branch_protection" "BranchProtect" {
	project = "${gitlab_project.foo.id}"
	branch = "BranchProtect-%d"
	push_access_level = "maintainer"
	merge_access_level = "maintainer"
}
	`, rInt, rInt)
}
