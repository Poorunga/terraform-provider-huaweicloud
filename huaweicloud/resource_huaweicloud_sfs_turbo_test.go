package huaweicloud

import (
	"fmt"
	"testing"

	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/shares"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
)

func TestAccSFSTurbo_basic(t *testing.T) {
	randSuffix := acctest.RandString(5)
	turboName := fmt.Sprintf("sfs-turbo-acc-%s", randSuffix)
	resourceName := "huaweicloud_sfs_turbo.sfs-turbo1"
	var turbo shares.Turbo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSFSTurboDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSFSTurbo_basic(randSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSFSTurboExists(resourceName, &turbo),
					resource.TestCheckResourceAttr(resourceName, "name", turboName),
					resource.TestCheckResourceAttr(resourceName, "share_proto", "NFS"),
					resource.TestCheckResourceAttr(resourceName, "share_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "enhanced", "false"),
					resource.TestCheckResourceAttr(resourceName, "size", "500"),
					resource.TestCheckResourceAttr(resourceName, "status", "200"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSFSTurbo_update(randSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSFSTurboExists(resourceName, &turbo),
					resource.TestCheckResourceAttr(resourceName, "size", "600"),
					resource.TestCheckResourceAttr(resourceName, "status", "221"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar_update"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value_update"),
				),
			},
		},
	})
}

func TestAccSFSTurbo_crypt(t *testing.T) {
	randSuffix := acctest.RandString(5)
	turboName := fmt.Sprintf("sfs-turbo-acc-%s", randSuffix)
	resourceName := "huaweicloud_sfs_turbo.sfs-turbo1"
	var turbo shares.Turbo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSFSTurboDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSFSTurbo_crypt(randSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSFSTurboExists(resourceName, &turbo),
					resource.TestCheckResourceAttr(resourceName, "name", turboName),
					resource.TestCheckResourceAttr(resourceName, "share_proto", "NFS"),
					resource.TestCheckResourceAttr(resourceName, "share_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "enhanced", "false"),
					resource.TestCheckResourceAttr(resourceName, "size", "500"),
					resource.TestCheckResourceAttr(resourceName, "status", "200"),
					resource.TestCheckResourceAttrSet(resourceName, "crypt_key_id"),
				),
			},
		},
	})
}

func TestAccSFSTurbo_withEpsId(t *testing.T) {
	randSuffix := acctest.RandString(5)
	turboName := fmt.Sprintf("sfs-turbo-acc-%s", randSuffix)
	resourceName := "huaweicloud_sfs_turbo.sfs-turbo1"
	var turbo shares.Turbo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckEpsID(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSFSTurboDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSFSTurbo_withEpsId(randSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSFSTurboExists(resourceName, &turbo),
					resource.TestCheckResourceAttr(resourceName, "name", turboName),
					resource.TestCheckResourceAttr(resourceName, "enterprise_project_id", HW_ENTERPRISE_PROJECT_ID_TEST),
				),
			},
		},
	})
}

func TestAccSFSTurbo_prePaid(t *testing.T) {
	randName := fmt.Sprintf("sfs-turbo-acc-%s", acctest.RandString(5))
	resourceName := "huaweicloud_sfs_turbo.test"
	var turbo shares.Turbo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckChargingMode(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSFSTurboDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSFSTurbo_prePaid_step1(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSFSTurboExists(resourceName, &turbo),
					resource.TestCheckResourceAttr(resourceName, "name", randName),
					resource.TestCheckResourceAttr(resourceName, "size", "500"),
				),
			},
			{
				Config: testAccSFSTurbo_prePaid_step2(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSFSTurboExists(resourceName, &turbo),
					resource.TestCheckResourceAttr(resourceName, "name", randName),
					resource.TestCheckResourceAttr(resourceName, "size", "600"),
				),
			},
		},
	})
}

func testAccCheckSFSTurboDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*config.Config)
	sfsClient, err := config.SfsV1Client(HW_REGION_NAME)
	if err != nil {
		return fmtp.Errorf("Error creating Huaweicloud sfs turbo client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "huaweicloud_sfs_turbo" {
			continue
		}

		_, err := shares.Get(sfsClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmtp.Errorf("SFS Turbo still exists")
		}
	}

	return nil
}

func testAccCheckSFSTurboExists(n string, share *shares.Turbo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmtp.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmtp.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*config.Config)
		sfsClient, err := config.SfsV1Client(HW_REGION_NAME)
		if err != nil {
			return fmtp.Errorf("Error creating Huaweicloud sfs turbo client: %s", err)
		}

		found, err := shares.Get(sfsClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmtp.Errorf("sfs turbo not found")
		}

		*share = *found
		return nil
	}
}

func testAccNetworkPreConditions(suffix string) string {
	return fmt.Sprintf(`
resource "huaweicloud_vpc_v1" "test" {
  name = "tf-acc-vpc-%s"
  cidr = "192.168.0.0/16"
}

resource "huaweicloud_vpc_subnet_v1" "test" {
  name       = "tf-acc-subnet-%s"
  cidr       = "192.168.0.0/16"
  gateway_ip = "192.168.0.1"
  vpc_id     = huaweicloud_vpc_v1.test.id
}

resource "huaweicloud_networking_secgroup_v2" "secgroup" {
  name        = "tf-acc-sg-%s"
  description = "terraform security group for sfs turbo acceptance test"
}
`, suffix, suffix, suffix)
}

func testAccSFSTurbo_basic(suffix string) string {
	return fmt.Sprintf(`
%s
data "huaweicloud_availability_zones" "myaz" {}

resource "huaweicloud_sfs_turbo" "sfs-turbo1" {
  name        = "sfs-turbo-acc-%s"
  size        = 500
  share_proto = "NFS"
  vpc_id      = huaweicloud_vpc_v1.test.id
  subnet_id   = huaweicloud_vpc_subnet_v1.test.id
  security_group_id = huaweicloud_networking_secgroup_v2.secgroup.id
  availability_zone = data.huaweicloud_availability_zones.myaz.names[0]

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, testAccNetworkPreConditions(suffix), suffix)
}

func testAccSFSTurbo_update(suffix string) string {
	return fmt.Sprintf(`
%s
data "huaweicloud_availability_zones" "myaz" {}

resource "huaweicloud_sfs_turbo" "sfs-turbo1" {
  name        = "sfs-turbo-acc-%s"
  size        = 600
  share_proto = "NFS"
  vpc_id      = huaweicloud_vpc_v1.test.id
  subnet_id   = huaweicloud_vpc_subnet_v1.test.id
  security_group_id = huaweicloud_networking_secgroup_v2.secgroup.id
  availability_zone = data.huaweicloud_availability_zones.myaz.names[0]

  tags = {
    foo = "bar_update"
    key = "value_update"
  }
}
`, testAccNetworkPreConditions(suffix), suffix)
}

func testAccSFSTurbo_crypt(suffix string) string {
	return fmt.Sprintf(`
%s
data "huaweicloud_availability_zones" "myaz" {}

resource "huaweicloud_kms_key" "key_1" {
  key_alias = "kms-acc-%s"
  pending_days = "7"
}

resource "huaweicloud_sfs_turbo" "sfs-turbo1" {
  name        = "sfs-turbo-acc-%s"
  size        = 500
  share_proto = "NFS"
  vpc_id      = huaweicloud_vpc_v1.test.id
  subnet_id   = huaweicloud_vpc_subnet_v1.test.id
  security_group_id = huaweicloud_networking_secgroup_v2.secgroup.id
  availability_zone = data.huaweicloud_availability_zones.myaz.names[0]
  crypt_key_id      = huaweicloud_kms_key.key_1.id
}
`, testAccNetworkPreConditions(suffix), suffix, suffix)
}

func testAccSFSTurbo_withEpsId(suffix string) string {
	return fmt.Sprintf(`
%s
data "huaweicloud_availability_zones" "myaz" {}

resource "huaweicloud_sfs_turbo" "sfs-turbo1" {
  name                   = "sfs-turbo-acc-%s"
  size                   = 500
  share_proto            = "NFS"
  vpc_id                 = huaweicloud_vpc_v1.test.id
  subnet_id              = huaweicloud_vpc_subnet_v1.test.id
  security_group_id      = huaweicloud_networking_secgroup_v2.secgroup.id
  availability_zone      = data.huaweicloud_availability_zones.myaz.names[0]
  enterprise_project_id  = "%s"
}
`, testAccNetworkPreConditions(suffix), suffix, HW_ENTERPRISE_PROJECT_ID_TEST)
}

func testAccSFSTurbo_prePaid_step1(name string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_sfs_turbo" "test" {
  vpc_id                 = huaweicloud_vpc.test.id
  subnet_id              = huaweicloud_vpc_subnet.test.id
  security_group_id      = huaweicloud_networking_secgroup.test.id
  availability_zone      = data.huaweicloud_availability_zones.test.names[0]
  name                   = "%[2]s"
  size                   = 500
  share_proto            = "NFS"
  enterprise_project_id  = "0"

  charging_mode = "prePaid"
  period_unit   = "month"
  period        = 1
  auto_renew    = false
}
`, common.TestBaseNetwork(name), name)
}

func testAccSFSTurbo_prePaid_step2(name string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_sfs_turbo" "test" {
  vpc_id                 = huaweicloud_vpc.test.id
  subnet_id              = huaweicloud_vpc_subnet.test.id
  security_group_id      = huaweicloud_networking_secgroup.test.id
  availability_zone      = data.huaweicloud_availability_zones.test.names[0]
  name                   = "%[2]s"
  size                   = 600
  share_proto            = "NFS"
  enterprise_project_id  = "0"

  charging_mode = "prePaid"
  period_unit   = "month"
  period        = 1
  auto_renew    = true
}
`, common.TestBaseNetwork(name), name)
}
