package ims

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/common/tags"
	"github.com/chnsz/golangsdk/openstack/imageservice/v2/images"
	"github.com/chnsz/golangsdk/openstack/ims/v1/imagecopy"
	"github.com/chnsz/golangsdk/openstack/ims/v2/cloudimages"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceImsImageCopy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImsImageCopyCreate,
		UpdateContext: resourceImsImageCopyUpdate,
		ReadContext:   resourceImsImageCopyRead,
		DeleteContext: resourceImsImageCopyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"source_image_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `Specifies the ID of the source image.`,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the name of the copy image.`,
			},
			"target_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the target region name.`,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the description of the copy image.`,
			},
			"kms_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the master key used for encrypting an image.`,
			},
			"enterprise_project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the enterprise project id of the image.`,
			},
			"agency_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the agency name.`,
			},
			"vault_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the ID of the vault.`,
			},
			"tags": common.TagsSchema(),
			// following are additional attributes
			"instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_size": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceImsImageCopyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	sourceRegion := cfg.GetRegion(d)

	var jobId string
	imsV1Client, err := cfg.ImageV1Client(sourceRegion)
	if err != nil {
		return diag.Errorf("error creating IMS client: %s", err)
	}

	imsV2Client, err := getImsV2Client(d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}

	targetRegion := d.Get("target_region").(string)
	if targetRegion == "" || targetRegion == sourceRegion {
		withinRegionCopyOpts := imagecopy.WithinRegionCopyOpts{
			Name:                d.Get("name").(string),
			Description:         d.Get("description").(string),
			CmkId:               d.Get("kms_key_id").(string),
			EnterpriseProjectID: common.GetEnterpriseProjectID(d, cfg),
		}

		log.Printf("[DEBUG] Within region copy Options: %#v", withinRegionCopyOpts)

		sourceImageId := d.Get("source_image_id").(string)
		jobRes, err := imagecopy.WithinRegionCopy(imsV1Client, sourceImageId, withinRegionCopyOpts).ExtractJobStatus()
		if err != nil {
			return diag.Errorf("error creating image copy within region: %s", err)
		}
		jobId = jobRes.JobID
	} else {
		crossRegionCopyOpts := imagecopy.CrossRegionCopyOpts{
			Name:              d.Get("name").(string),
			Description:       d.Get("description").(string),
			TargetRegion:      targetRegion,
			TargetProjectName: targetRegion,
			AgencyName:        d.Get("agency_name").(string),
			VaultId:           d.Get("vault_id").(string),
		}

		log.Printf("[DEBUG] Cross region copy Options: %#v", crossRegionCopyOpts)

		sourceImageId := d.Get("source_image_id").(string)
		jobRes, err := imagecopy.CrossRegionCopy(imsV1Client, sourceImageId, crossRegionCopyOpts).ExtractJobStatus()
		if err != nil {
			return diag.Errorf("error creating image copy cross region: %s", err)
		}
		jobId = jobRes.JobID
	}

	// Wait for the copy image to become available.
	log.Printf("[DEBUG] Waiting for IMS to become available")
	err = cloudimages.WaitForJobSuccess(imsV1Client, int(d.Timeout(schema.TimeoutCreate)/time.Second), jobId)
	if err != nil {
		return diag.FromErr(err)
	}

	imageId, err := cloudimages.GetJobEntity(imsV1Client, jobId, "image_id")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(imageId.(string))

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		tagList := utils.ExpandResourceTags(tagRaw)
		if tagErr := tags.Create(imsV2Client, "images", d.Id(), tagList).ExtractErr(); tagErr != nil {
			return diag.Errorf("error setting tags of images %s: %s", d.Id(), tagErr)
		}
	}

	return resourceImsImageCopyRead(ctx, d, meta)
}

func resourceImsImageCopyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	imsClient, err := getImsV2Client(d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") {
		updateOpts := make(images.UpdateOpts, 0)
		v := images.ReplaceImageName{NewName: d.Get("name").(string)}
		updateOpts = append(updateOpts, v)

		log.Printf("[DEBUG] Update Options: %#v", updateOpts)
		_, err = images.Update(imsClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return diag.Errorf("error updating image: %s", err)
		}
	}

	// update tags
	if d.HasChange("tags") {
		tagErr := utils.UpdateResourceTags(imsClient, d, "images", d.Id())
		if tagErr != nil {
			return diag.Errorf("error updating tags of IMS image :%s, err:%s", d.Id(), tagErr)
		}
	}

	return resourceImsImageCopyRead(ctx, d, meta)
}

func resourceImsImageCopyRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	imsClient, err := getImsV2Client(d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}

	img, err := GetCloudImage(imsClient, d.Id())
	if err != nil {
		return diag.Errorf("image %s not found: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Retrieved Image %s: %#v", d.Id(), img)

	mErr := multierror.Append(
		d.Set("region", region),
		d.Set("name", img.Name),
		d.Set("description", img.Description),
		d.Set("kms_key_id", img.SystemCmkid),
		d.Set("instance_id", getInstanceID(img.DataOrigin)),
		d.Set("os_version", img.OsVersion),
		d.Set("visibility", img.Visibility),
		d.Set("data_origin", img.DataOrigin),
		d.Set("disk_format", img.DiskFormat),
		d.Set("image_size", img.ImageSize),
		d.Set("checksum", img.Checksum),
		d.Set("status", img.Status),
		d.Set("enterprise_project_id", img.EnterpriseProjectID),
	)

	// fetch tags
	if resourceTags, err := tags.Get(imsClient, "image", d.Id()).Extract(); err == nil {
		tagMap := utils.TagsToMap(resourceTags.Tags)
		mErr = multierror.Append(mErr, d.Set("tags", tagMap))
	} else {
		log.Printf("[WARN] Fetching tags of IMS images failed: %s", err)
	}

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceImsImageCopyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	imsClient, err := getImsV2Client(d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting Image %s", d.Id())
	if err = images.Delete(imsClient, d.Id()).Err; err != nil {
		return diag.Errorf("error deleting image copy: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    waitForImageDelete(imsClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for delete image (%s) complete: %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

func getImsV2Client(d *schema.ResourceData, cfg *config.Config) (*golangsdk.ServiceClient, error) {
	imageRegion := cfg.GetRegion(d)
	if v, ok := d.GetOk("target_region"); ok {
		imageRegion = v.(string)
	}

	imsClient, err := cfg.ImageV2Client(imageRegion)
	if err != nil {
		return nil, fmt.Errorf("error creating IMS client: %s", err)
	}
	return imsClient, nil
}
