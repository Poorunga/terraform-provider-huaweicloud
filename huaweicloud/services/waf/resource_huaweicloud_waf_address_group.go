// ---------------------------------------------------------------
// *** AUTO GENERATED CODE ***
// @Product WAF
// ---------------------------------------------------------------

package waf

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceWafAddressGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddressGroupCreate,
		UpdateContext: resourceAddressGroupUpdate,
		ReadContext:   resourceAddressGroupRead,
		DeleteContext: resourceAddressGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the name of the address group.`,
			},
			"ip_addresses": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: `Specifies the IP addresses or IP address ranges.`,
			},
			"enterprise_project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `Specifies the enterprise project ID of WAF address group.`,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the description of the address group.`,
			},
			"rules": {
				Type:        schema.TypeList,
				Elem:        rulesSchema(),
				Computed:    true,
				Description: `The list of rules that use the IP address group.`,
			},
		},
	}
}

func rulesSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"rule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The ID of rule.`,
			},
			"rule_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The name of rule.`,
			},
			"policy_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The ID of policy.`,
			},
			"policy_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The name of policy.`,
			},
		},
	}
	return &sc
}

func resourceAddressGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	// createWAFAddressGroup: create WAF address group.
	var (
		createWAFAddressGroupHttpUrl = "v1/{project_id}/waf/ip-groups"
		createWAFAddressGroupProduct = "waf"
	)
	createWAFAddressGroupClient, err := cfg.NewServiceClient(createWAFAddressGroupProduct, region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	createWAFAddressGroupPath := createWAFAddressGroupClient.Endpoint + createWAFAddressGroupHttpUrl
	createWAFAddressGroupPath = strings.ReplaceAll(createWAFAddressGroupPath, "{project_id}",
		createWAFAddressGroupClient.ProjectID)
	createWAFAddressGroupPath += buildWAFAddressGroupQueryParams(d, cfg)

	createWAFAddressGroupOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	createWAFAddressGroupOpt.JSONBody = utils.RemoveNil(buildWAFAddressGroupBodyParams(d))
	createWAFAddressGroupResp, err := createWAFAddressGroupClient.Request("POST", createWAFAddressGroupPath,
		&createWAFAddressGroupOpt)
	if err != nil {
		return diag.Errorf("error creating address group: %s", err)
	}

	createWAFAddressGroupRespBody, err := utils.FlattenResponse(createWAFAddressGroupResp)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := jmespath.Search("id", createWAFAddressGroupRespBody)
	if err != nil {
		return diag.Errorf("error creating address group: ID is not found in API response")
	}
	d.SetId(id.(string))

	return resourceAddressGroupRead(ctx, d, meta)
}

func buildWAFAddressGroupBodyParams(d *schema.ResourceData) map[string]interface{} {
	ipAddresses := d.Get("ip_addresses").([]interface{})
	addresses := make([]string, 0, len(ipAddresses))
	for _, v := range ipAddresses {
		addresses = append(addresses, v.(string))
	}

	bodyParams := map[string]interface{}{
		"name":        utils.ValueIngoreEmpty(d.Get("name")),
		"ips":         strings.Join(addresses, ","),
		"description": utils.ValueIngoreEmpty(d.Get("description")),
	}
	return bodyParams
}

func buildWAFAddressGroupQueryParams(d *schema.ResourceData, conf *config.Config) string {
	epsId := common.GetEnterpriseProjectID(d, conf)
	if epsId == "" {
		return ""
	}
	return fmt.Sprintf("?enterprise_project_id=%s", epsId)
}

func resourceAddressGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var mErr *multierror.Error

	// getWAFAddressGroup: Query WAF address group
	var (
		getWAFAddressGroupHttpUrl = "v1/{project_id}/waf/ip-group/{id}"
		getWAFAddressGroupProduct = "waf"
	)
	getWAFAddressGroupClient, err := cfg.NewServiceClient(getWAFAddressGroupProduct, region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	getWAFAddressGroupPath := getWAFAddressGroupClient.Endpoint + getWAFAddressGroupHttpUrl
	getWAFAddressGroupPath = strings.ReplaceAll(getWAFAddressGroupPath, "{project_id}",
		getWAFAddressGroupClient.ProjectID)
	getWAFAddressGroupPath = strings.ReplaceAll(getWAFAddressGroupPath, "{id}", d.Id())
	getWAFAddressGroupPath += buildWAFAddressGroupQueryParams(d, cfg)

	getWAFAddressGroupOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	getWAFAddressGroupResp, err := getWAFAddressGroupClient.Request("GET", getWAFAddressGroupPath,
		&getWAFAddressGroupOpt)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving address group")
	}

	getWAFAddressGroupRespBody, err := utils.FlattenResponse(getWAFAddressGroupResp)
	if err != nil {
		return diag.FromErr(err)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("region", region),
		d.Set("name", utils.PathSearch("name", getWAFAddressGroupRespBody, nil)),
		d.Set("ip_addresses", flattenAddressGroupResponseBodyIpAddresses(getWAFAddressGroupRespBody)),
		d.Set("description", utils.PathSearch("description", getWAFAddressGroupRespBody, nil)),
		d.Set("rules", flattenAddressGroupResponseBodyRules(getWAFAddressGroupRespBody)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenAddressGroupResponseBodyIpAddresses(resp interface{}) []string {
	if resp == nil {
		return nil
	}
	ips := utils.PathSearch("ips", resp, "").(string)
	return strings.Split(ips, ",")
}

func flattenAddressGroupResponseBodyRules(resp interface{}) []interface{} {
	if resp == nil {
		return nil
	}
	curJson := utils.PathSearch("rules", resp, make([]interface{}, 0))
	curArray := curJson.([]interface{})
	rst := make([]interface{}, 0, len(curArray))
	for _, v := range curArray {
		rst = append(rst, map[string]interface{}{
			"rule_id":     utils.PathSearch("rule_id", v, nil),
			"rule_name":   utils.PathSearch("rule_name", v, nil),
			"policy_id":   utils.PathSearch("policy_id", v, nil),
			"policy_name": utils.PathSearch("policy_name", v, nil),
		})
	}
	return rst
}

func resourceAddressGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	updateWAFAddressGroupChanges := []string{
		"name",
		"ip_addresses",
		"description",
	}

	if d.HasChanges(updateWAFAddressGroupChanges...) {
		// updateWAFAddressGroup: Update WAF address group
		var (
			updateWAFAddressGroupHttpUrl = "v1/{project_id}/waf/ip-group/{id}"
			updateWAFAddressGroupProduct = "waf"
		)
		updateWAFAddressGroupClient, err := cfg.NewServiceClient(updateWAFAddressGroupProduct, region)
		if err != nil {
			return diag.Errorf("error creating WAF Client: %s", err)
		}

		updateWAFAddressGroupPath := updateWAFAddressGroupClient.Endpoint + updateWAFAddressGroupHttpUrl
		updateWAFAddressGroupPath = strings.ReplaceAll(updateWAFAddressGroupPath, "{project_id}",
			updateWAFAddressGroupClient.ProjectID)
		updateWAFAddressGroupPath = strings.ReplaceAll(updateWAFAddressGroupPath, "{id}", d.Id())
		updateWAFAddressGroupPath += buildWAFAddressGroupQueryParams(d, cfg)

		updateWAFAddressGroupOpt := golangsdk.RequestOpts{
			MoreHeaders: map[string]string{
				"Content-Type": "application/json;charset=utf8",
			},
			KeepResponseBody: true,
			OkCodes: []int{
				200,
			},
		}
		updateWAFAddressGroupOpt.JSONBody = utils.RemoveNil(buildWAFAddressGroupBodyParams(d))
		_, err = updateWAFAddressGroupClient.Request("PUT", updateWAFAddressGroupPath,
			&updateWAFAddressGroupOpt)
		if err != nil {
			return diag.Errorf("error updating address group: %s", err)
		}
	}
	return resourceAddressGroupRead(ctx, d, meta)
}

func resourceAddressGroupDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	// deleteWAFAddressGroup: Delete WAF address group
	var (
		deleteWAFAddressGroupHttpUrl = "v1/{project_id}/waf/ip-group/{id}"
		deleteWAFAddressGroupProduct = "waf"
	)
	deleteWAFAddressGroupClient, err := cfg.NewServiceClient(deleteWAFAddressGroupProduct, region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	deleteWAFAddressGroupPath := deleteWAFAddressGroupClient.Endpoint + deleteWAFAddressGroupHttpUrl
	deleteWAFAddressGroupPath = strings.ReplaceAll(deleteWAFAddressGroupPath, "{project_id}",
		deleteWAFAddressGroupClient.ProjectID)
	deleteWAFAddressGroupPath = strings.ReplaceAll(deleteWAFAddressGroupPath, "{id}", d.Id())
	deleteWAFAddressGroupPath += buildWAFAddressGroupQueryParams(d, cfg)

	deleteWAFAddressGroupOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	_, err = deleteWAFAddressGroupClient.Request("DELETE", deleteWAFAddressGroupPath, &deleteWAFAddressGroupOpt)
	if err != nil {
		return diag.Errorf("error deleting address group: %s", err)
	}

	return nil
}
