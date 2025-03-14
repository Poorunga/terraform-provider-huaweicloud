package dms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmespath/go-jmespath"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/common/tags"
	"github.com/chnsz/golangsdk/openstack/dms/v2/availablezones"
	"github.com/chnsz/golangsdk/openstack/dms/v2/kafka/instances"
	"github.com/chnsz/golangsdk/openstack/dms/v2/products"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceDmsKafkaInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsKafkaInstanceCreate,
		ReadContext:   resourceDmsKafkaInstanceRead,
		UpdateContext: resourceDmsKafkaInstanceUpdate,
		DeleteContext: resourceDmsKafkaInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(50 * time.Minute),
			Update: schema.DefaultTimeout(50 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_spec_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"manager_user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"manager_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"availability_zones": {
				// There is a problem with order of elements in Availability Zone list returned by Kafka API.
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "schema: Required",
			},
			"flavor_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"product_id"},
				RequiredWith: []string{"broker_num", "storage_space"},
			},
			"broker_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_space": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"access_user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				RequiredWith: []string{
					"password",
				},
			},
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
				ForceNew:  true,
				RequiredWith: []string{
					"access_user",
				},
			},
			"maintain_begin": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"maintain_end": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"public_ip_ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"retention_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"produce_reject", "time_base",
				}, false),
			},
			"dumping": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enable_auto_topic": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enterprise_project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": common.TagsSchema(),
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"partition_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enable_public_ip": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ssl_enable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"used_storage_space": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connect_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_spec_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"management_connect_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Typo, it is only kept in the code, will not be shown in the docs.
			"manegement_connect_address": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "typo in manegement_connect_address, please use \"management_connect_address\" instead.",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"available_zones": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"available_zones", "availability_zones"},
				Deprecated:   "available_zones has deprecated, please use \"availability_zones\" instead.",
			},
			"bandwidth": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Deprecated: "The bandwidth has been deprecated. " +
					"If you need to change the bandwidth, please update the product_id.",
			},
			"cross_vpc_accesses": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MinItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"advertised_ip": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"listener_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"port_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// Typo, it is only kept in the code, will not be shown in the docs.
						"lisenter_ip": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "typo in lisenter_ip, please use \"listener_ip\" instead.",
						},
					},
				},
			},
			"charging_mode": common.SchemaChargingMode(nil),
			"period_unit":   common.SchemaPeriodUnit(nil),
			"period":        common.SchemaPeriod(nil),
			"auto_renew":    common.SchemaAutoRenewUpdatable(nil),
		},
	}
}

func validateAndBuildPublicIpIDParam(publicIpIDs []interface{}, bandwidth string) (string, error) {
	bandwidthAndIPMapper := map[string]int{
		"100MB":  3,
		"300MB":  3,
		"600MB":  4,
		"1200MB": 8,
	}
	needIpCount := bandwidthAndIPMapper[bandwidth]

	if needIpCount != len(publicIpIDs) {
		return "", fmt.Errorf("error creating Kafka instance: "+
			"%d public ip IDs needed when bandwidth is set to %s, but got %d",
			needIpCount, bandwidth, len(publicIpIDs))
	}
	return strings.Join(utils.ExpandToStringList(publicIpIDs), ","), nil
}

func getKafkaProductDetails(cfg *config.Config, d *schema.ResourceData) (*products.Detail, error) {
	productRsp, err := getProducts(cfg, cfg.GetRegion(d), engineKafka)
	if err != nil {
		return nil, fmt.Errorf("error querying Kafka product list: %s", err)
	}

	productID := d.Get("product_id").(string)
	engineVersion := d.Get("engine_version").(string)

	for _, ps := range productRsp.Hourly {
		if ps.Version != engineVersion {
			continue
		}
		for _, v := range ps.Values {
			for _, p := range v.Details {
				if p.ProductID == productID {
					return &p, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("can not found Kafka product details base on product_id: %s", productID)
}

func updateCrossVpcAccess(client *golangsdk.ServiceClient, d *schema.ResourceData) error {
	newVal := d.Get("cross_vpc_accesses")
	var crossVpcAccessArr []map[string]interface{}

	instance, err := instances.Get(client, d.Id()).Extract()
	if err != nil {
		return fmt.Errorf("error getting DMS Kafka instance: %v", err)
	}

	crossVpcAccessArr, err = flattenCrossVpcInfo(instance.CrossVpcInfo)
	if err != nil {
		return fmt.Errorf("error retrieving details of the cross-VPC: %v", err)
	}

	newAccessArr := newVal.([]interface{})
	contentMap := make(map[string]string)
	for i, oldAccess := range crossVpcAccessArr {
		listenerIp := oldAccess["listener_ip"].(string)
		if listenerIp == "" {
			listenerIp = oldAccess["lisenter_ip"].(string)
		}
		// If we configure the advertised ip as ["192.168.0.19", "192.168.0.8"], the length of new accesses is 2,
		// and the length of old accesses is always 3.
		if len(newAccessArr) > i {
			// Make sure the index is valid.
			newAccess := newAccessArr[i].(map[string]interface{})
			// Since the "advertised_ip" is already a definition in the schema, the key name must exist.
			if advIp, ok := newAccess["advertised_ip"].(string); ok && advIp != "" {
				contentMap[listenerIp] = advIp
				continue
			}
		}
		contentMap[listenerIp] = listenerIp
	}

	log.Printf("[DEBUG} Update Kafka cross-vpc contentMap: %#v", contentMap)

	updateRst, err := instances.UpdateCrossVpc(client, d.Id(), instances.CrossVpcUpdateOpts{
		Contents: contentMap,
	})
	if err != nil {
		return fmt.Errorf("error updating advertised IP: %v", err)
	}

	if !updateRst.Success {
		failedIps := make([]string, 0, len(updateRst.Connections))
		for _, conn := range updateRst.Connections {
			if !conn.Success {
				failedIps = append(failedIps, conn.ListenersIp)
			}
		}
		return fmt.Errorf("failed to update the advertised IPs corresponding to some listener IPs (%v)", failedIps)
	}
	return nil
}

func resourceDmsKafkaInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	client, err := cfg.DmsV2Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error initializing DMS Kafka(v2) client: %s", err)
	}

	var dErr diag.Diagnostics
	if _, ok := d.GetOk("flavor_id"); ok {
		dErr = createKafkaInstanceWithFlavor(ctx, d, meta)
	} else {
		dErr = createKafkaInstanceWithProductID(ctx, d, meta)
	}
	if dErr != nil {
		return dErr
	}

	// After the kafka instance is created, wait for the access port to complete the binding.
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"PENDING"},
		Target:       []string{"BOUND"},
		Refresh:      portBindStatusRefreshFunc(client, d.Id(), 0),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		dErr = diag.Errorf("Kafka instance is created, but failed to enable cross-VPC %s : %s", d.Id(), err)
		dErr[0].Severity = diag.Warning
		return dErr
	}

	if _, ok := d.GetOk("cross_vpc_accesses"); ok {
		if err = updateCrossVpcAccess(client, d); err != nil {
			return diag.Errorf("failed to update default advertised IP: %s", err)
		}
	}

	return resourceDmsKafkaInstanceRead(ctx, d, meta)
}

func createKafkaInstanceWithFlavor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conf := meta.(*config.Config)
	region := conf.GetRegion(d)
	client, err := conf.DmsV2Client(region)
	if err != nil {
		return diag.Errorf("error initializing DMS Kafka(v2) client: %s", err)
	}

	createOpts := &instances.CreateOps{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Engine:              engineKafka,
		EngineVersion:       d.Get("engine_version").(string),
		AccessUser:          d.Get("access_user").(string),
		VPCID:               d.Get("vpc_id").(string),
		SecurityGroupID:     d.Get("security_group_id").(string),
		SubnetID:            d.Get("network_id").(string),
		ProductID:           d.Get("flavor_id").(string),
		KafkaManagerUser:    d.Get("manager_user").(string),
		MaintainBegin:       d.Get("maintain_begin").(string),
		MaintainEnd:         d.Get("maintain_end").(string),
		RetentionPolicy:     d.Get("retention_policy").(string),
		ConnectorEnalbe:     d.Get("dumping").(bool),
		EnableAutoTopic:     d.Get("enable_auto_topic").(bool),
		StorageSpecCode:     d.Get("storage_spec_code").(string),
		StorageSpace:        d.Get("storage_space").(int),
		BrokerNum:           d.Get("broker_num").(int),
		EnterpriseProjectID: common.GetEnterpriseProjectID(d, conf),
	}

	if chargingMode, ok := d.GetOk("charging_mode"); ok && chargingMode == "prePaid" {
		var autoRenew bool
		if d.Get("auto_renew").(string) == "true" {
			autoRenew = true
		}
		isAutoPay := true
		createOpts.BssParam = instances.BssParam{
			ChargingMode: d.Get("charging_mode").(string),
			PeriodType:   d.Get("period_unit").(string),
			PeriodNum:    d.Get("period").(int),
			IsAutoRenew:  &autoRenew,
			IsAutoPay:    &isAutoPay,
		}
	}

	if ids, ok := d.GetOk("public_ip_ids"); ok {
		createOpts.EnablePublicIP = true
		createOpts.PublicIpID = strings.Join(utils.ExpandToStringList(ids.([]interface{})), ",")
	}

	createOpts.SslEnable = false
	if d.Get("access_user").(string) != "" && d.Get("password").(string) != "" {
		createOpts.SslEnable = true
	}

	var availableZones []string
	if zoneIDs, ok := d.GetOk("available_zones"); ok {
		availableZones = utils.ExpandToStringList(zoneIDs.([]interface{}))
	} else {
		// convert the codes of the availability zone into ids
		azCodes := d.Get("availability_zones").(*schema.Set)
		availableZones, err = getAvailableZoneIDByCode(conf, region, azCodes.List())
		if err != nil {
			return diag.FromErr(err)
		}
	}
	createOpts.AvailableZones = availableZones

	// set tags
	if tagRaw := d.Get("tags").(map[string]interface{}); len(tagRaw) > 0 {
		createOpts.Tags = utils.ExpandResourceTags(tagRaw)
	}
	log.Printf("[DEBUG] Create DMS Kafka instance options: %#v", createOpts)
	// Add password here, so it wouldn't go in the above log entry
	createOpts.Password = d.Get("password").(string)
	createOpts.KafkaManagerPassword = d.Get("manager_password").(string)

	kafkaInstance, err := instances.Create(client, createOpts).Extract()
	if err != nil {
		return diag.Errorf("error creating Kafka instance: %s", err)
	}
	instanceID := kafkaInstance.InstanceID

	var delayTime time.Duration = 300
	if chargingMode, ok := d.GetOk("charging_mode"); ok && chargingMode == "prePaid" {
		err = waitForOrderComplete(ctx, d, conf, client, instanceID)
		if err != nil {
			return diag.FromErr(err)
		}
		delayTime = 5
	}

	log.Printf("[INFO] Creating Kafka instance, ID: %s", instanceID)
	d.SetId(instanceID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"CREATING"},
		Target:       []string{"RUNNING"},
		Refresh:      KafkaInstanceStateRefreshFunc(client, instanceID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        delayTime * time.Second,
		PollInterval: 15 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for Kafka instance (%s) to be ready: %s", instanceID, err)
	}

	return nil
}

func createKafkaInstanceWithProductID(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)
	client, err := cfg.DmsV2Client(region)
	if err != nil {
		return diag.Errorf("error initializing DMS Kafka(v2) client: %s", err)
	}

	product, err := getKafkaProductDetails(cfg, d)
	if err != nil {
		return diag.Errorf("Error querying product detail: %s", err)
	}

	bandwidth := product.Bandwidth
	defaultPartitionNum, _ := strconv.ParseInt(product.PartitionNum, 10, 64)
	defaultStorageSpace, _ := strconv.ParseInt(product.Storage, 10, 64)

	// check storage
	storageSpace, ok := d.GetOk("storage_space")
	if ok && storageSpace.(int) < int(defaultStorageSpace) {
		return diag.Errorf("storage capacity (storage_space) must be greater than the minimum capacity of the product, "+
			"product capacity is %v, got: %v", defaultStorageSpace, storageSpace)
	}

	sslEnable := false
	if d.Get("access_user").(string) != "" && d.Get("password").(string) != "" {
		sslEnable = true
	}

	var availableZones []string
	if zoneIDs, ok := d.GetOk("available_zones"); ok {
		availableZones = utils.ExpandToStringList(zoneIDs.([]interface{}))
	} else {
		// Convert AZ Codes to AZ IDs
		azCodes := d.Get("availability_zones").(*schema.Set)
		availableZones, err = getAvailableZoneIDByCode(cfg, region, azCodes.List())
		if err != nil {
			return diag.FromErr(err)
		}
	}

	createOpts := &instances.CreateOps{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Engine:              engineKafka,
		EngineVersion:       d.Get("engine_version").(string),
		Specification:       bandwidth,
		StorageSpace:        int(defaultStorageSpace),
		PartitionNum:        int(defaultPartitionNum),
		AccessUser:          d.Get("access_user").(string),
		VPCID:               d.Get("vpc_id").(string),
		SecurityGroupID:     d.Get("security_group_id").(string),
		SubnetID:            d.Get("network_id").(string),
		AvailableZones:      availableZones,
		ProductID:           d.Get("product_id").(string),
		KafkaManagerUser:    d.Get("manager_user").(string),
		MaintainBegin:       d.Get("maintain_begin").(string),
		MaintainEnd:         d.Get("maintain_end").(string),
		SslEnable:           sslEnable,
		RetentionPolicy:     d.Get("retention_policy").(string),
		ConnectorEnalbe:     d.Get("dumping").(bool),
		EnableAutoTopic:     d.Get("enable_auto_topic").(bool),
		StorageSpecCode:     d.Get("storage_spec_code").(string),
		EnterpriseProjectID: common.GetEnterpriseProjectID(d, cfg),
	}

	if chargingMode, ok := d.GetOk("charging_mode"); ok && chargingMode == "prePaid" {
		var autoRenew bool
		if d.Get("auto_renew").(string) == "true" {
			autoRenew = true
		}
		isAutoPay := true
		createOpts.BssParam = instances.BssParam{
			ChargingMode: d.Get("charging_mode").(string),
			PeriodType:   d.Get("period_unit").(string),
			PeriodNum:    d.Get("period").(int),
			IsAutoRenew:  &autoRenew,
			IsAutoPay:    &isAutoPay,
		}
	}

	if pubIpIDs, ok := d.GetOk("public_ip_ids"); ok {
		publicIpIDs, err := validateAndBuildPublicIpIDParam(pubIpIDs.([]interface{}), bandwidth)
		if err != nil {
			return diag.FromErr(err)
		}
		createOpts.EnablePublicIP = true
		createOpts.PublicIpID = publicIpIDs
	}

	// set tags
	if tagsRaw := d.Get("tags").(map[string]interface{}); len(tagsRaw) > 0 {
		createOpts.Tags = utils.ExpandResourceTags(tagsRaw)
	}
	log.Printf("[DEBUG] Create DMS Kafka instance options: %#v", createOpts)

	// Add password here, so it wouldn't go in the above log entry
	createOpts.Password = d.Get("password").(string)
	createOpts.KafkaManagerPassword = d.Get("manager_password").(string)

	kafkaInstance, err := instances.Create(client, createOpts).Extract()
	if err != nil {
		return diag.Errorf("error creating DMS kafka instance: %s", err)
	}
	instanceID := kafkaInstance.InstanceID

	var delayTime time.Duration = 300
	if chargingMode, ok := d.GetOk("charging_mode"); ok && chargingMode == "prePaid" {
		err = waitForOrderComplete(ctx, d, cfg, client, instanceID)
		if err != nil {
			return diag.FromErr(err)
		}
		delayTime = 5
	}

	log.Printf("[INFO] Creating Kafka instance, ID: %s", instanceID)

	// Store the instance ID now
	d.SetId(instanceID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"CREATING"},
		Target:       []string{"RUNNING"},
		Refresh:      KafkaInstanceStateRefreshFunc(client, instanceID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        delayTime * time.Second,
		PollInterval: 15 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Kafka instance (%s) to be ready: %s", instanceID, err)
	}

	// resize storage capacity of the instance
	if ok && storageSpace.(int) != int(defaultStorageSpace) {
		err = resizeKafkaInstanceStorage(ctx, d, client)
		if err != nil {
			dErrs := diag.Errorf("Kafka instance is created, but fails resize the storage capacity, "+
				"expected %v GB, but got %v GB, error: %s ", storageSpace.(int), defaultStorageSpace, err)
			dErrs[0].Severity = diag.Warning
			return dErrs
		}
	}

	return nil
}

func waitForOrderComplete(ctx context.Context, d *schema.ResourceData, conf *config.Config,
	client *golangsdk.ServiceClient, instanceID string) error {
	region := conf.GetRegion(d)
	orderId, err := getInstanceOrderId(ctx, d, client, instanceID)
	if err != nil {
		return err
	}
	if orderId == "" {
		log.Printf("[WARN] error get order id by instance ID: %s", instanceID)
		return nil
	}

	bssClient, err := conf.BssV2Client(region)
	if err != nil {
		return fmt.Errorf("error creating BSS v2 client: %s", err)
	}
	// wait for order success
	err = common.WaitOrderComplete(ctx, bssClient, orderId, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}
	_, err = common.WaitOrderResourceComplete(ctx, bssClient, orderId, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error waiting for Kafka order resource %s complete: %s", orderId, err)
	}
	return nil
}

func getInstanceOrderId(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient,
	instanceID string) (string, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"EMPTY"},
		Target:       []string{"CREATING"},
		Refresh:      kafkaInstanceCreatingFunc(client, instanceID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        500 * time.Millisecond,
		PollInterval: 500 * time.Millisecond,
	}
	orderId, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return "", fmt.Errorf("error waiting for Kafka instance (%s) to creating: %s", instanceID, err)
	}
	return orderId.(string), nil
}

func kafkaInstanceCreatingFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res := instances.Get(client, instanceID)
		if res.Err != nil {
			actualCode, err := jmespath.Search("Actual", res.Err)
			if err != nil {
				return nil, "", err
			}
			if actualCode == 404 {
				return res, "EMPTY", nil
			}
		}
		instance, err := res.Extract()
		if err != nil {
			return nil, "", err
		}
		return instance.OrderID, "CREATING", nil
	}
}

func flattenCrossVpcInfo(crossVpcInfoStr string) ([]map[string]interface{}, error) {
	if crossVpcInfoStr == "" {
		return nil, nil
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] Recover panic when flattening cross-VPC infos structure: %#v", r)
		}
	}()

	crossVpcInfos := make(map[string]interface{})
	err := json.Unmarshal([]byte(crossVpcInfoStr), &crossVpcInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal CrossVpcInfo, crossVpcInfo: %s, error: %s", crossVpcInfoStr, err)
	}

	ipArr := make([]string, 0, len(crossVpcInfos))
	for ip := range crossVpcInfos {
		ipArr = append(ipArr, ip)
	}
	sort.Strings(ipArr) // Sort by listeners IP.

	result := make([]map[string]interface{}, len(crossVpcInfos))
	for i, ip := range ipArr {
		crossVpcInfo := crossVpcInfos[ip].(map[string]interface{})
		result[i] = map[string]interface{}{
			"listener_ip":   ip,
			"lisenter_ip":   ip,
			"advertised_ip": crossVpcInfo["advertised_ip"],
			"port":          crossVpcInfo["port"],
			"port_id":       crossVpcInfo["port_id"],
		}
	}
	return result, nil
}

func setKafkaFlavorId(d *schema.ResourceData, flavorId string) error {
	re := regexp.MustCompile(`^[a-z0-9]+\.\d+u\d+g\.cluster|single$`)
	if re.MatchString(flavorId) {
		return d.Set("flavor_id", flavorId)
	}
	return d.Set("product_id", flavorId)
}

func resourceDmsKafkaInstanceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	client, err := cfg.DmsV2Client(region)
	if err != nil {
		return diag.Errorf("error initializing DMS Kafka(v2) client: %s", err)
	}

	v, err := instances.Get(client, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "DMS Kafka instance")
	}
	log.Printf("[DEBUG] Get Kafka instance: %+v", v)

	crossVpcAccess, err := flattenCrossVpcInfo(v.CrossVpcInfo)
	if err != nil {
		return diag.Errorf("error parsing the cross-VPC information: %v", err)
	}

	partitionNum, _ := strconv.ParseInt(v.PartitionNum, 10, 64)
	// Convert the AZ ids to AZ codes.
	availableZoneIDs := v.AvailableZones
	availableZoneCodes, err := getAvailableZoneCodeByID(cfg, region, availableZoneIDs)
	mErr := multierror.Append(nil, err)

	var chargingMode = "postPaid"
	if v.ChargingMode == 0 {
		chargingMode = "prePaid"
	}

	mErr = multierror.Append(mErr,
		d.Set("region", cfg.GetRegion(d)),
		setKafkaFlavorId(d, v.ProductID), // Set flavor_id or product_id.
		d.Set("name", v.Name),
		d.Set("description", v.Description),
		d.Set("engine", v.Engine),
		d.Set("engine_version", v.EngineVersion),
		d.Set("bandwidth", v.Specification),
		// storage_space indicates total_storage_space while creating
		// set value of total_storage_space to storage_space to keep consistent
		d.Set("storage_space", v.TotalStorageSpace),
		d.Set("partition_num", partitionNum),
		d.Set("vpc_id", v.VPCID),
		d.Set("security_group_id", v.SecurityGroupID),
		d.Set("network_id", v.SubnetID),
		d.Set("available_zones", availableZoneIDs),
		d.Set("availability_zones", availableZoneCodes),
		d.Set("broker_num", v.BrokerNum),
		d.Set("manager_user", v.KafkaManagerUser),
		d.Set("maintain_begin", v.MaintainBegin),
		d.Set("maintain_end", v.MaintainEnd),
		d.Set("enable_public_ip", v.EnablePublicIP),
		d.Set("ssl_enable", v.SslEnable),
		d.Set("retention_policy", v.RetentionPolicy),
		d.Set("dumping", v.ConnectorEnalbe),
		d.Set("enable_auto_topic", v.EnableAutoTopic),
		d.Set("storage_spec_code", v.StorageSpecCode),
		d.Set("enterprise_project_id", v.EnterpriseProjectID),
		d.Set("used_storage_space", v.UsedStorageSpace),
		d.Set("connect_address", v.ConnectAddress),
		d.Set("port", v.Port),
		d.Set("status", v.Status),
		d.Set("resource_spec_code", v.ResourceSpecCode),
		d.Set("user_id", v.UserID),
		d.Set("user_name", v.UserName),
		d.Set("manegement_connect_address", v.ManagementConnectAddress),
		d.Set("management_connect_address", v.ManagementConnectAddress),
		d.Set("type", v.Type),
		d.Set("access_user", v.AccessUser),
		d.Set("cross_vpc_accesses", crossVpcAccess),
		d.Set("charging_mode", chargingMode),
	)

	// set tags
	if resourceTags, err := tags.Get(client, engineKafka, d.Id()).Extract(); err == nil {
		tagMap := utils.TagsToMap(resourceTags.Tags)
		if err = d.Set("tags", tagMap); err != nil {
			mErr = multierror.Append(mErr,
				fmt.Errorf("error saving tags to state for DMS kafka instance (%s): %s", d.Id(), err))
		}
	} else {
		log.Printf("[WARN] error fetching tags of DMS kafka instance (%s): %s", d.Id(), err)
	}

	if mErr.ErrorOrNil() != nil {
		return diag.Errorf("failed to set attributes for DMS kafka instance: %s", mErr)
	}

	return nil
}

func resourceDmsKafkaInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	client, err := cfg.DmsV2Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error initializing DMS Kafka(v2) client: %s", err)
	}

	var mErr *multierror.Error
	if d.HasChanges("name", "description", "maintain_begin", "maintain_end",
		"security_group_id", "retention_policy", "enterprise_project_id") {
		description := d.Get("description").(string)
		updateOpts := instances.UpdateOpts{
			Description:         &description,
			MaintainBegin:       d.Get("maintain_begin").(string),
			MaintainEnd:         d.Get("maintain_end").(string),
			SecurityGroupID:     d.Get("security_group_id").(string),
			RetentionPolicy:     d.Get("retention_policy").(string),
			EnterpriseProjectID: d.Get("enterprise_project_id").(string),
		}

		if d.HasChange("name") {
			updateOpts.Name = d.Get("name").(string)
		}

		err = instances.Update(client, d.Id(), updateOpts).Err
		if err != nil {
			mErr = multierror.Append(mErr, fmt.Errorf("error updating Kafka Instance: %s", err))
		}
	}

	if d.HasChanges("product_id", "flavor_id", "storage_space", "broker_num") {
		err = resizeKafkaInstance(ctx, d, meta)
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	if d.HasChange("tags") {
		// update tags
		if err = utils.UpdateResourceTags(client, d, engineKafka, d.Id()); err != nil {
			mErr = multierror.Append(mErr, fmt.Errorf("error updating tags of Kafka instance: %s, err: %s",
				d.Id(), err))
		}
	}

	if d.HasChange("cross_vpc_accesses") {
		if err = updateCrossVpcAccess(client, d); err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	if d.HasChange("auto_renew") {
		bssClient, err := cfg.BssV2Client(cfg.GetRegion(d))
		if err != nil {
			return diag.Errorf("error creating BSS V2 client: %s", err)
		}
		if err = common.UpdateAutoRenew(bssClient, d.Get("auto_renew").(string), d.Id()); err != nil {
			return diag.Errorf("error updating the auto-renew of the Kafka instance (%s): %s", d.Id(), err)
		}
	}

	if mErr.ErrorOrNil() != nil {
		return diag.Errorf("error while updating DMS Kafka instances, %s", mErr)
	}
	return resourceDmsKafkaInstanceRead(ctx, d, meta)
}

func resizeKafkaInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	cfg := meta.(*config.Config)
	client, err := cfg.DmsV2Client(cfg.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error initializing DMS(v2) client: %s", err)
	}

	if d.HasChanges("product_id") {
		product, err := getKafkaProductDetails(cfg, d)
		if err != nil {
			return fmt.Errorf("failed to resize Kafka instance, query product details error: %s", err)
		}
		storageSpace := d.Get("storage_space").(int)
		resizeOpts := instances.ResizeInstanceOpts{
			NewSpecCode:     &product.SpecCode,
			NewStorageSpace: &storageSpace,
		}
		log.Printf("[DEBUG] Resize Kafka instance storage space options: %s", utils.MarshalValue(resizeOpts))

		if err := doKafkaInstanceResize(ctx, d, client, resizeOpts); err != nil {
			return err
		}
	}

	if d.HasChanges("flavor_id") {
		flavorID := d.Get("flavor_id").(string)
		operType := "vertical"
		resizeOpts := instances.ResizeInstanceOpts{
			OperType:     &operType,
			NewProductID: &flavorID,
		}
		log.Printf("[DEBUG] Resize Kafka instance flavor ID options: %s", utils.MarshalValue(resizeOpts))

		if err := doKafkaInstanceResize(ctx, d, client, resizeOpts); err != nil {
			return err
		}
	}

	if d.HasChanges("broker_num") {
		operType := "horizontal"
		brokerNum := d.Get("broker_num").(int)

		resizeOpts := instances.ResizeInstanceOpts{
			OperType:     &operType,
			NewBrokerNum: &brokerNum,
		}
		log.Printf("[DEBUG] Resize Kafka instance broker number options: %s", utils.MarshalValue(resizeOpts))

		if err := doKafkaInstanceResize(ctx, d, client, resizeOpts); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:      []string{"PENDING"},
			Target:       []string{"BOUND"},
			Refresh:      portBindStatusRefreshFunc(client, d.Id(), brokerNum),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			Delay:        10 * time.Second,
			PollInterval: 10 * time.Second,
		}
		if _, err = stateConf.WaitForStateContext(ctx); err != nil {
			return err
		}
	}

	if d.HasChanges("storage_space") {
		if err = resizeKafkaInstanceStorage(ctx, d, client); err != nil {
			return err
		}
	}

	return nil
}

func resizeKafkaInstanceStorage(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient) error {
	newStorageSpace := d.Get("storage_space").(int)
	operType := "storage"
	resizeOpts := instances.ResizeInstanceOpts{
		OperType:        &operType,
		NewStorageSpace: &newStorageSpace,
	}
	log.Printf("[DEBUG] Resize Kafka instance storage space options: %s", utils.MarshalValue(resizeOpts))

	return doKafkaInstanceResize(ctx, d, client, resizeOpts)
}

func doKafkaInstanceResize(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient, opts instances.ResizeInstanceOpts) error {
	if _, err := instances.Resize(client, d.Id(), opts); err != nil {
		return fmt.Errorf("resize Kafka instance failed: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"PENDING"},
		Target:       []string{"RUNNING"},
		Refresh:      kafkaResizeStateRefresh(client, d, opts.OperType),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        180 * time.Second,
		PollInterval: 15 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for instance (%s) to resize: %v", d.Id(), err)
	}
	return nil
}

func kafkaResizeStateRefresh(client *golangsdk.ServiceClient, d *schema.ResourceData, operType *string) resource.StateRefreshFunc {
	flavorID := d.Get("flavor_id").(string)
	if flavorID == "" {
		flavorID = d.Get("product_id").(string)
	}
	storageSpace := d.Get("storage_space").(int)
	brokerNum := d.Get("broker_num").(int)

	return func() (interface{}, string, error) {
		v, err := instances.Get(client, d.Id()).Extract()
		if err != nil {
			return nil, "failed", err
		}

		if ((operType == nil || *operType == "vertical") && v.ProductID != flavorID) ||
			(operType != nil && *operType == "storage" && v.TotalStorageSpace != storageSpace) ||
			(operType != nil && *operType == "horizontal" && v.BrokerNum != brokerNum) {
			return v, "PENDING", nil
		}

		return v, v.Status, nil
	}
}

func resourceDmsKafkaInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	client, err := cfg.DmsV2Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error initializing DMS Kafka(v2) client: %s", err)
	}

	if d.Get("charging_mode") == "prePaid" {
		if err = common.UnsubscribePrePaidResource(d, cfg, []string{d.Id()}); err != nil {
			return diag.Errorf("error unsubscribe Kafka instance: %s", err)
		}
	} else {
		err = instances.Delete(client, d.Id()).ExtractErr()
		if err != nil {
			return common.CheckDeletedDiag(d, err, "failed to delete Kafka instance")
		}
	}

	// Wait for the instance to delete before moving on.
	log.Printf("[DEBUG] Waiting for Kafka instance (%s) to be deleted", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"DELETING", "RUNNING", "ERROR"}, // Status may change to ERROR on deletion.
		Target:       []string{"DELETED"},
		Refresh:      KafkaInstanceStateRefreshFunc(client, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        120 * time.Second,
		PollInterval: 15 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for DMS Kafka instance (%s) to be deleted: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] DMS Kafka instance %s has been deleted", d.Id())
	d.SetId("")
	return nil
}

func portBindStatusRefreshFunc(client *golangsdk.ServiceClient, instanceID string, brokerNum int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := instances.Get(client, instanceID).Extract()
		if err != nil {
			return nil, "QUERY ERROR", err
		}
		if brokerNum == 0 && resp.CrossVpcInfo != "" {
			return resp, "BOUND", nil
		}
		if brokerNum != 0 && resp.CrossVpcInfo != "" {
			crossVpcInfoMap, err := flattenCrossVpcInfo(resp.CrossVpcInfo)
			if err != nil {
				return resp, "ParseError", err
			}

			if len(crossVpcInfoMap) == brokerNum {
				return resp, "BOUND", nil
			}
		}
		return resp, "PENDING", nil
	}
}

func KafkaInstanceStateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := instances.Get(client, instanceID).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return v, "DELETED", nil
			}
			return nil, "QUERY ERROR", err
		}

		return v, v.Status, nil
	}
}

func getAvailableZoneIDByCode(config *config.Config, region string, azCodes []interface{}) ([]string, error) {
	if len(azCodes) == 0 {
		return nil, fmt.Errorf(`arguments "azCodes" is required`)
	}

	availableZones, err := getAvailableZones(config, region)
	if err != nil {
		return nil, err
	}

	codeIDMapping := make(map[string]string)
	for _, v := range availableZones {
		codeIDMapping[v.Code] = v.ID
	}

	azIDs := make([]string, 0, len(azCodes))
	for _, code := range azCodes {
		if id, ok := codeIDMapping[code.(string)]; ok {
			azIDs = append(azIDs, id)
		}
	}
	log.Printf("[DEBUG] DMS converts the AZ codes to AZ IDs: \n%#v => \n%#v", azCodes, azIDs)
	return azIDs, nil
}

func getAvailableZoneCodeByID(config *config.Config, region string, azIDs []string) ([]string, error) {
	if len(azIDs) == 0 {
		return nil, fmt.Errorf(`arguments "azIDs" is required`)
	}

	availableZones, err := getAvailableZones(config, region)
	if err != nil {
		return nil, err
	}

	idCodeMapping := make(map[string]string)
	for _, v := range availableZones {
		idCodeMapping[v.ID] = v.Code
	}

	azCodes := make([]string, 0, len(azIDs))
	for _, id := range azIDs {
		if code, ok := idCodeMapping[id]; ok {
			azCodes = append(azCodes, code)
		}
	}
	log.Printf("[DEBUG] DMS converts the AZ IDs to AZ codes: \n%#v => \n%#v", azIDs, azCodes)
	return azCodes, nil
}

func getAvailableZones(cfg *config.Config, region string) ([]availablezones.AvailableZone, error) {
	client, err := cfg.DmsV2Client(region)
	if err != nil {
		return nil, fmt.Errorf("error initializing DMS(v2) client: %s", err)
	}

	r, err := availablezones.Get(client)
	if err != nil {
		return nil, fmt.Errorf("error querying available Zones: %s", err)
	}

	return r.AvailableZones, nil
}
