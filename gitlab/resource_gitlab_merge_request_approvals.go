package gitlab

import (
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabMergeRequestApprovals() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabMergeRequestApprovalsChange,
		Update: resourceGitlabMergeRequestApprovalsChange,
		Read:   resourceGitlabMergeRequestApprovalsRead,
		Delete: resourceGitlabMergeRequestApprovalsReset,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"reset_approvals_on_push": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},
			"disable_overriding_approvers_per_merge_request": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},
			"merge_requests_author_approval": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceGitlabMergeRequestApprovalsChange(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	resetApprovalsOnPush := d.Get("reset_approvals_on_push").(bool)
	disableOverridingApprovers := d.Get("disable_overriding_approvers_per_merge_request").(bool)
	mergeRequestsAuthorsApproval := d.Get("merge_requests_author_approval").(bool)

	options := &gitlab.ChangeApprovalConfigurationOptions{
		ResetApprovalsOnPush:                      &resetApprovalsOnPush,
		DisableOverridingApproversPerMergeRequest: &disableOverridingApprovers,
		MergeRequestsAuthorApproval:               &mergeRequestsAuthorsApproval,
	}

	log.Printf("[DEBUG] Change gitlab project %s approvals %#v", project, *options)

	approval, _, err := client.Projects.ChangeApprovalConfiguration(project, options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", approval.ID))

	return resourceGitlabMergeRequestApprovalsRead(d, meta)
}

func resourceGitlabMergeRequestApprovalsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	log.Printf("[DEBUG] Get gitlab project %s approvals", project)

	approvals, _, err := client.Projects.GetApprovalConfiguration(project)
	if err != nil {
		return err
	}

	d.Set("project", project)
	d.Set("reset_approvals_on_push", approvals.ResetApprovalsOnPush)
	d.Set("disable_overriding_approvers_per_merge_request", approvals.DisableOverridingApproversPerMergeRequest)
	d.Set("merge_requests_author_approval", approvals.MergeRequestsAuthorApproval)

	return nil
}

func resourceGitlabMergeRequestApprovalsReset(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	log.Printf("[DEBUG] Reset gitlab project %s approvals", project)

	options := &gitlab.ChangeApprovalConfigurationOptions{
		ResetApprovalsOnPush:                      false,
		DisableOverridingApproversPerMergeRequest: false,
		MergeRequestsAuthorApproval:               false,
	}
	_, _, err := client.Projects.ChangeApprovalConfiguration(project, options)
	return err
}
