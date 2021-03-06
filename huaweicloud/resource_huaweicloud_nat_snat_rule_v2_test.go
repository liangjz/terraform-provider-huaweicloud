package huaweicloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/hw_snatrules"
)

func TestAccNatSnatRule_basic(t *testing.T) {
	randSuffix := acctest.RandString(5)
	resourceName := "huaweicloud_nat_snat_rule.snat_1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckNat(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNatV2SnatRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNatV2SnatRule_basic(randSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatV2GatewayExists("huaweicloud_nat_gateway.nat_1"),
					testAccCheckNatV2SnatRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckNatV2SnatRuleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	natClient, err := config.natV2Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating HuaweiCloud nat client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "huaweicloud_nat_snat_rule" {
			continue
		}

		_, err := hw_snatrules.Get(natClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Snat rule still exists")
		}
	}

	return nil
}

func testAccCheckNatV2SnatRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		natClient, err := config.natV2Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating HuaweiCloud nat client: %s", err)
		}

		found, err := hw_snatrules.Get(natClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Snat rule not found")
		}

		return nil
	}
}

func testAccNatV2SnatRule_basic(suffix string) string {
	return fmt.Sprintf(`
%s

resource "huaweicloud_vpc_eip" "eip_1" {
  publicip {
    type = "5_bgp"
  }
  bandwidth {
    name        = "test"
    size        = 5
    share_type  = "PER"
    charge_mode = "traffic"
  }
}

resource "huaweicloud_nat_gateway" "nat_1" {
  name                = "nat-gateway-basic-%s"
  description         = "test for terraform"
  spec                = "1"
  internal_network_id = huaweicloud_vpc_subnet.subnet_1.id
  router_id           = huaweicloud_vpc.vpc_1.id
}

resource "huaweicloud_nat_snat_rule" "snat_1" {
  nat_gateway_id = huaweicloud_nat_gateway.nat_1.id
  network_id     = huaweicloud_vpc_subnet.subnet_1.id
  floating_ip_id = huaweicloud_vpc_eip.eip_1.id
}
	`, testAccNatPreCondition(suffix), suffix)
}
