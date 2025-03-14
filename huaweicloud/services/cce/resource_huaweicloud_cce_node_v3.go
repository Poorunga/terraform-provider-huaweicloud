package cce

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/cce/v3/clusters"
	"github.com/chnsz/golangsdk/openstack/cce/v3/nodes"
	"github.com/chnsz/golangsdk/openstack/common/tags"
	"github.com/chnsz/golangsdk/openstack/ecs/v1/cloudservers"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

func ResourceCCENodeV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCENodeV3Create,
		ReadContext:   resourceCCENodeV3Read,
		UpdateContext: resourceCCENodeV3Update,
		DeleteContext: resourceCCENodeV3Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCCENodeV3Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"os": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"key_pair": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"password", "key_pair"},
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"password", "key_pair"},
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"root_volume": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"volumetype": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"hw_passthrough": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "schema: Internal",
						},
						"extend_param": {
							Type:       schema.TypeString,
							Optional:   true,
							ForceNew:   true,
							Deprecated: "use extend_params instead",
						},
						"extend_params": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					}},
			},
			"data_volumes": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"volumetype": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"hw_passthrough": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "schema: Internal",
						},
						"extend_param": {
							Type:       schema.TypeString,
							Optional:   true,
							ForceNew:   true,
							Deprecated: "use extend_params instead",
						},
						"extend_params": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					}},
			},
			"storage": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"selectors": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  "evs",
									},
									"match_label_size": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"match_label_volume_type": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"match_label_metadata_encrypted": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"match_label_metadata_cmkid": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"match_label_count": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"groups": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"cce_managed": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},
									"selector_names": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"virtual_spaces": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"size": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
												},
												"lvm_lv_type": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												"lvm_path": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												"runtime_lv_type": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										}},
								},
							},
						},
					}},
			},
			"taints": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"effect": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					}},
			},
			"eip_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ConflictsWith: []string{
					"eip_ids", "iptype", "bandwidth_charge_mode", "bandwidth_size", "sharetype",
				},
			},
			"iptype": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				RequiredWith: []string{
					"iptype", "bandwidth_size", "sharetype",
				},
			},
			"bandwidth_charge_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"eip_ids", "eip_id"},
			},
			"sharetype": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				RequiredWith: []string{
					"iptype", "bandwidth_size", "sharetype",
				},
			},
			"bandwidth_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				RequiredWith: []string{
					"iptype", "bandwidth_size", "sharetype",
				},
			},
			"runtime": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"docker", "containerd",
				}, false),
			},
			"ecs_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ecs_performance_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "schema: Internal",
			},
			"product_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "schema: Internal",
			},
			"max_pods": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "schema: Internal",
			},
			"preinstall": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				StateFunc: utils.DecodeHashAndHexEncode,
			},
			"postinstall": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				StateFunc: utils.DecodeHashAndHexEncode,
			},
			"labels": { //(k8s_tags)
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			//(node/ecs_tags)
			"tags": common.TagsSchema(),
			"annotations": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "schema: Internal",
			},

			// charge info: charging_mode, period_unit, period, auto_renew, auto_pay
			"charging_mode": common.SchemaChargingMode(nil),
			"period_unit":   common.SchemaPeriodUnit(nil),
			"period":        common.SchemaPeriod(nil),
			"auto_renew":    common.SchemaAutoRenewUpdatable(nil),
			"auto_pay":      common.SchemaAutoPay(nil),

			"extend_param": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"keep_ecs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "schema: Internal",
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Deprecated
			"eip_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				ConflictsWith: []string{
					"eip_id", "iptype", "bandwidth_charge_mode", "bandwidth_size", "sharetype",
				},
				Deprecated: "use eip_id instead",
			},
			"billing_mode": {
				Type:       schema.TypeInt,
				Optional:   true,
				ForceNew:   true,
				Computed:   true,
				Deprecated: "use charging_mode instead",
			},
			"extend_param_charging_mode": {
				Type:       schema.TypeInt,
				Optional:   true,
				ForceNew:   true,
				Deprecated: "use charging_mode instead",
			},
			"order_id": {
				Type:       schema.TypeString,
				Optional:   true,
				ForceNew:   true,
				Deprecated: "will be removed after v1.26.0",
			},
		},
	}
}

func resourceCCENodeAnnotationsV2(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("annotations").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func resourceCCENodeK8sTags(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("labels").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func resourceCCENodeTags(d *schema.ResourceData) []tags.ResourceTag {
	tagRaw := d.Get("tags").(map[string]interface{})
	return utils.ExpandResourceTags(tagRaw)
}

func resourceCCERootVolume(d *schema.ResourceData) nodes.VolumeSpec {
	var root nodes.VolumeSpec
	volumeRaw := d.Get("root_volume").([]interface{})
	if len(volumeRaw) == 1 {
		rawMap := volumeRaw[0].(map[string]interface{})
		root.Size = rawMap["size"].(int)
		root.VolumeType = rawMap["volumetype"].(string)
		root.HwPassthrough = rawMap["hw_passthrough"].(bool)
		root.ExtendParam = rawMap["extend_params"].(map[string]interface{})

		if rawMap["kms_key_id"].(string) != "" {
			metadata := nodes.VolumeMetadata{
				SystemEncrypted: "1",
				SystemCmkid:     rawMap["kms_key_id"].(string),
			}
			root.Metadata = &metadata
		}
	}

	return root
}

func resourceCCEDataVolume(d *schema.ResourceData) []nodes.VolumeSpec {
	volumeRaw := d.Get("data_volumes").([]interface{})
	volumes := make([]nodes.VolumeSpec, len(volumeRaw))
	for i, raw := range volumeRaw {
		rawMap := raw.(map[string]interface{})
		volumes[i] = nodes.VolumeSpec{
			Size:          rawMap["size"].(int),
			VolumeType:    rawMap["volumetype"].(string),
			HwPassthrough: rawMap["hw_passthrough"].(bool),
			ExtendParam:   rawMap["extend_params"].(map[string]interface{}),
		}
		if rawMap["kms_key_id"].(string) != "" {
			metadata := nodes.VolumeMetadata{
				SystemEncrypted: "1",
				SystemCmkid:     rawMap["kms_key_id"].(string),
			}
			volumes[i].Metadata = &metadata
		}
	}
	return volumes
}

func resourceCCETaint(d *schema.ResourceData) []nodes.TaintSpec {
	taintRaw := d.Get("taints").([]interface{})
	taints := make([]nodes.TaintSpec, len(taintRaw))
	for i, raw := range taintRaw {
		rawMap := raw.(map[string]interface{})
		taints[i] = nodes.TaintSpec{
			Key:    rawMap["key"].(string),
			Value:  rawMap["value"].(string),
			Effect: rawMap["effect"].(string),
		}
	}
	return taints
}

func resourceCCEEipIDs(d *schema.ResourceData) []string {
	if v, ok := d.GetOk("eip_id"); ok {
		return []string{v.(string)}
	}
	rawID := d.Get("eip_ids").(*schema.Set)
	id := make([]string, rawID.Len())
	for i, raw := range rawID.List() {
		id[i] = raw.(string)
	}
	return id
}

func resourceCCEExtendParam(d *schema.ResourceData) map[string]interface{} {
	extendParam := make(map[string]interface{})
	if v, ok := d.GetOk("extend_param"); ok {
		for key, val := range v.(map[string]interface{}) {
			extendParam[key] = val.(string)
		}
		if v, ok := extendParam["periodNum"]; ok {
			periodNum, err := strconv.Atoi(v.(string))
			if err != nil {
				logp.Printf("[WARNING] PeriodNum %s invalid, Type conversion error: %s", v.(string), err)
			}
			extendParam["periodNum"] = periodNum
		}
	}

	// assemble the charge info
	var isPrePaid bool
	var billingMode int

	if v, ok := d.GetOk("charging_mode"); ok && v.(string) == "prePaid" {
		isPrePaid = true
	}
	if v, ok := d.GetOk("billing_mode"); ok {
		billingMode = v.(int)
	}
	if isPrePaid || billingMode == 2 {
		extendParam["chargingMode"] = 2
		extendParam["isAutoRenew"] = "false"
		extendParam["isAutoPay"] = common.GetAutoPay(d)
	}

	if v, ok := d.GetOk("period_unit"); ok {
		extendParam["periodType"] = v.(string)
	}
	if v, ok := d.GetOk("period"); ok {
		extendParam["periodNum"] = v.(int)
	}
	if v, ok := d.GetOk("auto_renew"); ok {
		extendParam["isAutoRenew"] = v.(string)
	}

	if v, ok := d.GetOk("ecs_performance_type"); ok {
		extendParam["ecs:performancetype"] = v.(string)
	}
	if v, ok := d.GetOk("max_pods"); ok {
		extendParam["maxPods"] = v.(int)
	}
	if v, ok := d.GetOk("order_id"); ok {
		extendParam["orderID"] = v.(string)
	}
	if v, ok := d.GetOk("product_id"); ok {
		extendParam["productID"] = v.(string)
	}
	if v, ok := d.GetOk("public_key"); ok {
		extendParam["publicKey"] = v.(string)
	}
	if v, ok := d.GetOk("preinstall"); ok {
		extendParam["alpha.cce/preInstall"] = utils.TryBase64EncodeString(v.(string))
	}
	if v, ok := d.GetOk("postinstall"); ok {
		extendParam["alpha.cce/postInstall"] = utils.TryBase64EncodeString(v.(string))
	}

	return extendParam
}

func resourceCCEStorage(d *schema.ResourceData) *nodes.StorageSpec {
	if v, ok := d.GetOk("storage"); ok {
		var storageSpec nodes.StorageSpec
		storageSpecRaw := v.([]interface{})
		storageSpecRawMap := storageSpecRaw[0].(map[string]interface{})
		storageSelectorSpecRaw := storageSpecRawMap["selectors"].([]interface{})
		storageGroupSpecRaw := storageSpecRawMap["groups"].([]interface{})

		var selectors []nodes.StorageSelectorsSpec
		for _, s := range storageSelectorSpecRaw {
			var selector nodes.StorageSelectorsSpec
			sMap := s.(map[string]interface{})
			selector.Name = sMap["name"].(string)
			selector.StorageType = sMap["type"].(string)
			selector.MatchLabels.Size = sMap["match_label_size"].(string)
			selector.MatchLabels.VolumeType = sMap["match_label_volume_type"].(string)
			selector.MatchLabels.MetadataEncrypted = sMap["match_label_metadata_encrypted"].(string)
			selector.MatchLabels.MetadataCmkid = sMap["match_label_metadata_cmkid"].(string)
			selector.MatchLabels.Count = sMap["match_label_count"].(string)

			selectors = append(selectors, selector)
		}
		storageSpec.StorageSelectors = selectors

		var groups []nodes.StorageGroupsSpec
		for _, g := range storageGroupSpecRaw {
			var group nodes.StorageGroupsSpec
			gMap := g.(map[string]interface{})
			group.Name = gMap["name"].(string)
			group.CceManaged = gMap["cce_managed"].(bool)

			selectorNamesRaw := gMap["selector_names"].([]interface{})
			selectorNames := make([]string, 0, len(selectorNamesRaw))
			for _, v := range selectorNamesRaw {
				selectorNames = append(selectorNames, v.(string))
			}
			group.SelectorNames = selectorNames

			virtualSpacesRaw := gMap["virtual_spaces"].([]interface{})
			virtualSpaces := make([]nodes.VirtualSpacesSpec, 0, len(virtualSpacesRaw))
			for _, v := range virtualSpacesRaw {
				var virtualSpace nodes.VirtualSpacesSpec
				virtualSpaceMap := v.(map[string]interface{})
				virtualSpace.Name = virtualSpaceMap["name"].(string)
				virtualSpace.Size = virtualSpaceMap["size"].(string)
				if virtualSpaceMap["lvm_lv_type"].(string) != "" {
					var lvmConfig nodes.LVMConfigSpec
					lvmConfig.LvType = virtualSpaceMap["lvm_lv_type"].(string)
					lvmConfig.Path = virtualSpaceMap["lvm_path"].(string)
					virtualSpace.LVMConfig = &lvmConfig
				}
				if virtualSpaceMap["runtime_lv_type"].(string) != "" {
					var runtimeConfig nodes.RuntimeConfigSpec
					runtimeConfig.LvType = virtualSpaceMap["runtime_lv_type"].(string)
					virtualSpace.RuntimeConfig = &runtimeConfig
				}

				virtualSpaces = append(virtualSpaces, virtualSpace)
			}
			group.VirtualSpaces = virtualSpaces

			groups = append(groups, group)
		}

		storageSpec.StorageGroups = groups
		return &storageSpec
	}
	return nil
}

func resourceCCENodeV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	nodeClient, err := config.CceV3Client(config.GetRegion(d))
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud CCE Node client: %s", err)
	}

	// validation
	billingMode := 0
	if d.Get("charging_mode").(string) == "prePaid" || d.Get("billing_mode").(int) == 2 {
		billingMode = 2
		if err := common.ValidatePrePaidChargeInfo(d); err != nil {
			return diag.FromErr(err)
		}
	}
	// eipCount must be specified when bandwidth_size parameters was set
	eipCount := 0
	if _, ok := d.GetOk("bandwidth_size"); ok {
		eipCount = 1
	}

	// wait for the cce cluster to become available
	clusterid := d.Get("cluster_id").(string)
	stateCluster := &resource.StateChangeConf{
		Target:       []string{"Available"},
		Refresh:      waitForClusterAvailable(nodeClient, clusterid),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        5 * time.Second,
		PollInterval: 5 * time.Second,
	}
	_, err = stateCluster.WaitForStateContext(ctx)
	if err != nil {
		return fmtp.DiagErrorf("Error waiting for HuaweiCloud CCE cluster to be Available: %s", err)
	}

	createOpts := nodes.CreateOpts{
		Kind:       "Node",
		ApiVersion: "v3",
		Metadata: nodes.CreateMetaData{
			Name:        d.Get("name").(string),
			Annotations: resourceCCENodeAnnotationsV2(d),
		},
		Spec: nodes.Spec{
			Flavor:      d.Get("flavor_id").(string),
			Az:          d.Get("availability_zone").(string),
			Os:          d.Get("os").(string),
			RootVolume:  resourceCCERootVolume(d),
			DataVolumes: resourceCCEDataVolume(d),
			Storage:     resourceCCEStorage(d),
			PublicIP: nodes.PublicIPSpec{
				Ids:   resourceCCEEipIDs(d),
				Count: eipCount,
				Eip: nodes.EipSpec{
					IpType: d.Get("iptype").(string),
					Bandwidth: nodes.BandwidthOpts{
						ChargeMode: d.Get("bandwidth_charge_mode").(string),
						Size:       d.Get("bandwidth_size").(int),
						ShareType:  d.Get("sharetype").(string),
					},
				},
			},
			BillingMode: billingMode,
			Count:       1,
			NodeNicSpec: nodes.NodeNicSpec{
				PrimaryNic: nodes.PrimaryNic{
					SubnetId: d.Get("subnet_id").(string),
				},
			},
			EcsGroupID:  d.Get("ecs_group_id").(string),
			ExtendParam: resourceCCEExtendParam(d),
			Taints:      resourceCCETaint(d),
			K8sTags:     resourceCCENodeK8sTags(d),
			UserTags:    resourceCCENodeTags(d),
		},
	}

	if v, ok := d.GetOk("fixed_ip"); ok {
		createOpts.Spec.NodeNicSpec.PrimaryNic.FixedIps = []string{v.(string)}
	}
	if v, ok := d.GetOk("runtime"); ok {
		createOpts.Spec.RunTime = &nodes.RunTimeSpec{
			Name: v.(string),
		}
	}

	logp.Printf("[DEBUG] Create Options: %#v", createOpts)
	// Add loginSpec here so it wouldn't go in the above log entry
	var loginSpec nodes.LoginSpec
	if common.HasFilledOpt(d, "key_pair") {
		loginSpec = nodes.LoginSpec{
			SshKey: d.Get("key_pair").(string),
		}
	} else {
		password, err := utils.TryPasswordEncrypt(d.Get("password").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		loginSpec = nodes.LoginSpec{
			UserPassword: nodes.UserPassword{
				Username: "root",
				Password: password,
			},
		}
	}
	createOpts.Spec.Login = loginSpec

	s, err := nodes.Create(nodeClient, clusterid, createOpts).Extract()
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud Node: %s", err)
	}

	if orderId, ok := s.Spec.ExtendParam["orderID"]; ok && orderId != "" {
		bssClient, err := config.BssV2Client(config.GetRegion(d))
		if err != nil {
			return diag.Errorf("error creating BSS v2 client: %s", err)
		}
		err = common.WaitOrderComplete(ctx, bssClient, orderId.(string), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
		// The resource ID generated by the CBC service only means that the underlying ECS of CCE node is created.
		_, err = common.WaitOrderResourceComplete(ctx, bssClient, orderId.(string), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// The completion of the creation of the underlying resource (ECS) corresponding to the CCE node does not mean that
	// the creation of the CCE node is completed.
	nodeID, err := getResourceIDFromJob(ctx, nodeClient, s.Status.JobID, "CreateNode", "CreateNodeVM",
		d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(nodeID)

	logp.Printf("[DEBUG] Waiting for CCE Node (%s) to become available", s.Metadata.Name)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Build", "Installing"},
		Target:       []string{"Active"},
		Refresh:      waitForCceNodeActive(nodeClient, clusterid, nodeID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        20 * time.Second,
		PollInterval: 20 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud CCE Node: %s", err)
	}

	return resourceCCENodeV3Read(ctx, d, meta)
}

func resourceCCENodeV3Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	nodeClient, err := config.CceV3Client(config.GetRegion(d))
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud CCE Node client: %s", err)
	}
	clusterid := d.Get("cluster_id").(string)
	s, err := nodes.Get(nodeClient, clusterid, d.Id()).Extract()

	if err != nil {
		return common.CheckDeletedDiag(d, err, "Error retrieving HuaweiCloud CCE Node")
	}

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("name", s.Metadata.Name),
		d.Set("flavor_id", s.Spec.Flavor),
		d.Set("availability_zone", s.Spec.Az),
		d.Set("os", s.Spec.Os),
		d.Set("key_pair", s.Spec.Login.SshKey),
		d.Set("subnet_id", s.Spec.NodeNicSpec.PrimaryNic.SubnetId),
		d.Set("ecs_group_id", s.Spec.EcsGroupID),
	)

	if s.Spec.BillingMode != 0 {
		mErr = multierror.Append(mErr, d.Set("charging_mode", "prePaid"))
	}
	if s.Spec.RunTime != nil {
		mErr = multierror.Append(mErr, d.Set("runtime", s.Spec.RunTime.Name))
	}

	var volumes []map[string]interface{}
	for _, pairObject := range s.Spec.DataVolumes {
		volume := make(map[string]interface{})
		volume["size"] = pairObject.Size
		volume["volumetype"] = pairObject.VolumeType
		volume["hw_passthrough"] = pairObject.HwPassthrough
		volume["extend_params"] = pairObject.ExtendParam
		volume["extend_param"] = ""
		if pairObject.Metadata != nil {
			volume["kms_key_id"] = pairObject.Metadata.SystemCmkid
		}
		volumes = append(volumes, volume)
	}
	mErr = multierror.Append(mErr, d.Set("data_volumes", volumes))

	rootVolume := []map[string]interface{}{
		{
			"size":           s.Spec.RootVolume.Size,
			"volumetype":     s.Spec.RootVolume.VolumeType,
			"hw_passthrough": s.Spec.RootVolume.HwPassthrough,
			"extend_params":  s.Spec.RootVolume.ExtendParam,
			"extend_param":   "",
		},
	}
	if s.Spec.RootVolume.Metadata != nil {
		rootVolume[0]["kms_key_id"] = s.Spec.RootVolume.Metadata.SystemCmkid
	}
	mErr = multierror.Append(mErr, d.Set("root_volume", rootVolume))

	// set computed attributes
	serverId := s.Status.ServerID
	mErr = multierror.Append(mErr,
		d.Set("server_id", serverId),
		d.Set("private_ip", s.Status.PrivateIP),
		d.Set("public_ip", s.Status.PublicIP),
		d.Set("status", s.Status.Phase),
	)

	computeClient, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud compute client: %s", err)
	}

	// fetch key_pair from ECS instance
	if server, err := cloudservers.Get(computeClient, serverId).Extract(); err == nil {
		mErr = multierror.Append(mErr, d.Set("key_pair", server.KeyName))
	} else {
		logp.Printf("[WARN] Error fetching ECS instance (%s): %s", serverId, err)
	}

	// fetch tags from ECS instance
	if resourceTags, err := tags.Get(computeClient, "cloudservers", serverId).Extract(); err == nil {
		tagmap := utils.TagsToMap(resourceTags.Tags)
		mErr = multierror.Append(mErr, d.Set("tags", tagmap))
	} else {
		logp.Printf("[WARN] Error fetching tags of ECS instance (%s): %s", serverId, err)
	}

	if err = mErr.ErrorOrNil(); err != nil {
		return fmtp.DiagErrorf("Error setting CCE Node fields: %s", err)
	}
	return nil
}

func resourceCCENodeV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	nodeClient, err := config.CceV3Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud CCE client: %s", err)
	}
	computeClient, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud compute client: %s", err)
	}

	if d.HasChange("name") {
		var updateOpts nodes.UpdateOpts
		updateOpts.Metadata.Name = d.Get("name").(string)

		clusterid := d.Get("cluster_id").(string)
		_, err = nodes.Update(nodeClient, clusterid, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmtp.DiagErrorf("Error updating HuaweiCloud cce node: %s", err)
		}
	}

	serverId := d.Get("server_id").(string)

	// update node tags with ECS API
	if d.HasChange("tags") {
		tagErr := utils.UpdateResourceTags(computeClient, d, "cloudservers", serverId)
		if tagErr != nil {
			return fmtp.DiagErrorf("Error updating tags of cce node %s: %s", d.Id(), tagErr)
		}
	}

	// update node key_pair with DEW API
	if d.HasChange("key_pair") {
		kmsClient, err := config.KmsV3Client(region)
		if err != nil {
			return diag.Errorf("error creating KMS v3 client: %s", err)
		}

		currentPwd, _ := d.GetChange("password")
		o, n := d.GetChange("key_pair")
		keyPairOpts := &common.KeypairAuthOpts{
			InstanceID:       serverId,
			InUsedKeyPair:    o.(string),
			NewKeyPair:       n.(string),
			InUsedPrivateKey: d.Get("private_key").(string),
			Password:         currentPwd.(string),
			DisablePassword:  true,
			Timeout:          d.Timeout(schema.TimeoutUpdate),
		}
		if err := common.UpdateEcsInstanceKeyPair(ctx, computeClient, kmsClient, keyPairOpts); err != nil {
			return diag.FromErr(err)
		}
	}

	// update node password with ECS API
	// A new password takes effect after the ECS is started or restarted.
	if d.HasChange("password") {
		// if the password is empty, it means that the ECS instance will bind a new keypair
		if newPwd, ok := d.GetOk("password"); ok {
			err := cloudservers.ChangeAdminPassword(computeClient, serverId, newPwd.(string)).ExtractErr()
			if err != nil {
				return diag.Errorf("error changing password of cce node %s: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("auto_renew") {
		bssClient, err := config.BssV2Client(config.GetRegion(d))
		if err != nil {
			return diag.Errorf("error creating BSS V2 client: %s", err)
		}
		if err = common.UpdateAutoRenew(bssClient, d.Get("auto_renew").(string), d.Get("server_id").(string)); err != nil {
			// Do not output the underlying ECS instance ID externally.
			return diag.Errorf("error updating the auto-renew of the node (%s): %s", d.Id(), err)
		}
	}

	return resourceCCENodeV3Read(ctx, d, meta)
}

func resourceCCENodeV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	nodeClient, err := config.CceV3Client(config.GetRegion(d))
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud CCE client: %s", err)
	}

	clusterid := d.Get("cluster_id").(string)
	// remove node without deleting ecs
	if d.Get("keep_ecs").(bool) {
		var removeOpts nodes.RemoveOpts

		var loginSpec nodes.LoginSpec
		if common.HasFilledOpt(d, "key_pair") {
			loginSpec = nodes.LoginSpec{
				SshKey: d.Get("key_pair").(string),
			}
		} else {
			password, err := utils.TryPasswordEncrypt(d.Get("password").(string))
			if err != nil {
				return diag.FromErr(err)
			}
			loginSpec = nodes.LoginSpec{
				UserPassword: nodes.UserPassword{
					Username: "root",
					Password: password,
				},
			}
		}
		removeOpts.Spec.Login = loginSpec

		nodeItem := nodes.NodeItem{
			Uid: d.Id(),
		}
		removeOpts.Spec.Nodes = append(removeOpts.Spec.Nodes, nodeItem)

		err = nodes.Remove(nodeClient, clusterid, removeOpts).ExtractErr()
		if err != nil {
			return fmtp.DiagErrorf("Error removing HuaweiCloud CCE node: %s", err)
		}
	} else {
		// for prePaid node, firstly, we should unsubscribe the ecs server, and then delete it
		if d.Get("charging_mode").(string) == "prePaid" || d.Get("billing_mode").(int) == 2 {
			serverID := d.Get("server_id").(string)
			publicIP := d.Get("public_ip").(string)

			resourceIDs := make([]string, 0, 2)
			computeClient, err := config.ComputeV1Client(config.GetRegion(d))
			if err != nil {
				return fmtp.DiagErrorf("Error creating HuaweiCloud compute client: %s", err)
			}

			// check whether the ecs server of the perPaid exists before unsubscribe it
			// because resource could not be found cannot be unsubscribed
			if serverID != "" {
				server, err := cloudservers.Get(computeClient, serverID).Extract()
				if err != nil {
					return common.CheckDeletedDiag(d, err, "error retrieving compute instance")
				}

				if server.Status != "DELETED" && server.Status != "SOFT_DELETED" {
					resourceIDs = append(resourceIDs, serverID)
				}
			}

			// unsubscribe the eip if necessary
			if _, ok := d.GetOk("iptype"); ok && publicIP != "" {
				eipClient, err := config.NetworkingV1Client(config.GetRegion(d))
				if err != nil {
					return fmtp.DiagErrorf("Error creating networking client: %s", err)
				}

				epsID := "all_granted_eps"
				if eipID, err := common.GetEipIDbyAddress(eipClient, publicIP, epsID); err == nil {
					resourceIDs = append(resourceIDs, eipID)
				} else {
					logp.Printf("[WARN] Error fetching EIP ID of CCE Node (%s): %s", d.Id(), err)
				}
			}

			if len(resourceIDs) > 0 {
				if err := common.UnsubscribePrePaidResource(d, config, resourceIDs); err != nil {
					return fmtp.DiagErrorf("Error unsubscribing HuaweiCloud CCE node: %s", err)
				}
			}

			// wait for the ecs server of the prePaid node to be deleted
			pending := []string{"ACTIVE", "SHUTOFF"}
			target := []string{"DELETED", "SOFT_DELETED"}
			deleteTimeout := d.Timeout(schema.TimeoutDelete)
			if err := waitForServerTargetState(ctx, computeClient, serverID, pending, target, deleteTimeout); err != nil {
				return fmtp.DiagErrorf("State waiting timeout: %s", err)
			}

			nodes.Delete(nodeClient, clusterid, d.Id())
		} else {
			err = nodes.Delete(nodeClient, clusterid, d.Id()).ExtractErr()
			if err != nil {
				return fmtp.DiagErrorf("Error deleting HuaweiCloud CCE node: %s", err)
			}
		}
	}
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Deleting"},
		Target:       []string{"Deleted"},
		Refresh:      waitForCceNodeDelete(nodeClient, clusterid, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        60 * time.Second,
		PollInterval: 20 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmtp.DiagErrorf("Error deleting HuaweiCloud CCE Node: %s", err)
	}

	d.SetId("")
	return nil
}

func getResourceIDFromJob(ctx context.Context, client *golangsdk.ServiceClient, jobID, jobType, subJobType string,
	timeout time.Duration) (string, error) {

	stateJob := &resource.StateChangeConf{
		Pending:      []string{"Initializing", "Running"},
		Target:       []string{"Success"},
		Refresh:      waitForJobStatus(client, jobID),
		Timeout:      timeout,
		Delay:        120 * time.Second,
		PollInterval: 20 * time.Second,
	}

	v, err := stateJob.WaitForStateContext(ctx)
	if err != nil {
		if job, ok := v.(*nodes.Job); ok {
			return "", fmtp.Errorf("Error waiting for job (%s) to become success: %s, reason: %s",
				jobID, err, job.Status.Reason)
		}

		return "", fmtp.Errorf("Error waiting for job (%s) to become success: %s", jobID, err)
	}

	job := v.(*nodes.Job)
	if len(job.Spec.SubJobs) == 0 {
		return "", fmtp.Errorf("Error fetching sub jobs from %s", jobID)
	}

	var subJobID string
	var refreshJob bool
	for _, s := range job.Spec.SubJobs {
		// postPaid: should get details of sub job ID
		if s.Spec.Type == jobType {
			subJobID = s.Metadata.ID
			refreshJob = true
			break
		}
	}

	if refreshJob {
		job, err = nodes.GetJobDetails(client, subJobID).ExtractJob()
		if err != nil {
			return "", fmtp.Errorf("Error fetching sub Job %s: %s", subJobID, err)
		}
	}

	var nodeid string
	for _, s := range job.Spec.SubJobs {
		if s.Spec.Type == subJobType {
			nodeid = s.Spec.ResourceID
			break
		}
	}
	if nodeid == "" {
		return "", fmtp.Errorf("Error fetching %s Job resource id", subJobType)
	}
	return nodeid, nil
}

func waitForCceNodeActive(cceClient *golangsdk.ServiceClient, clusterId, nodeId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := nodes.Get(cceClient, clusterId, nodeId).Extract()
		if err != nil {
			return nil, "", err
		}

		return n, n.Status.Phase, nil
	}
}

func waitForCceNodeDelete(cceClient *golangsdk.ServiceClient, clusterId, nodeId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		logp.Printf("[DEBUG] Attempting to delete HuaweiCloud CCE Node %s", nodeId)

		r, err := nodes.Get(cceClient, clusterId, nodeId).Extract()

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				logp.Printf("[DEBUG] Successfully deleted HuaweiCloud CCE Node %s", nodeId)
				return r, "Deleted", nil
			}
			return r, "Deleting", err
		}

		return r, r.Status.Phase, nil
	}
}

func waitForClusterAvailable(cceClient *golangsdk.ServiceClient, clusterId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		logp.Printf("[INFO] Waiting for CCE Cluster %s to be available", clusterId)
		n, err := clusters.Get(cceClient, clusterId).Extract()

		if err != nil {
			return nil, "", err
		}

		return n, n.Status.Phase, nil
	}
}

func waitForJobStatus(cceClient *golangsdk.ServiceClient, jobID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		job, err := nodes.GetJobDetails(cceClient, jobID).ExtractJob()
		if err != nil {
			return nil, "", err
		}

		return job, job.Status.Phase, nil
	}
}

func resourceCCENodeV3Import(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		err := fmtp.Errorf("Invalid format specified for CCE Node. Format must be <cluster id>/<node id>")
		return nil, err
	}

	clusterID := parts[0]
	nodeID := parts[1]

	d.SetId(nodeID)
	d.Set("cluster_id", clusterID)

	return []*schema.ResourceData{d}, nil
}

func waitForServerTargetState(ctx context.Context, client *golangsdk.ServiceClient, id string,
	pending, target []string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:      pending,
		Target:       target,
		Refresh:      ServerV1StateRefreshFunc(client, id),
		Timeout:      timeout,
		Delay:        5 * time.Second,
		PollInterval: 5 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmtp.Errorf("error waiting for instance (%s) to become target state (%v): %s", id, target, err)
	}
	return nil
}

// ServerV1StateRefreshFunc returns a resource.StateRefreshFunc that is used to watch an HuaweiCloud instance.
func ServerV1StateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		s, err := cloudservers.Get(client, instanceID).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return s, "DELETED", nil
			}
			return nil, "", err
		}

		// get fault message when status is ERROR
		if s.Status == "ERROR" {
			fault := fmtp.Errorf("[error code: %d, message: %s]", s.Fault.Code, s.Fault.Message)
			return s, "ERROR", fault
		}
		return s, s.Status, nil
	}
}
