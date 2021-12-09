package g42vbcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetworkingV2PortDataSource_basic(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2PortDataSource_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.g42vbcloud_networking_port.port_3", "all_fixed_ips.#", "1"),
				),
			},
		},
	})
}

func testAccNetworkingV2PortDataSource_basic() string {
	return fmt.Sprintf(`
data "g42vbcloud_vpc_subnet" "mynet" {
  name = "subnet-default"
}

data "g42vbcloud_networking_port" "port_3" {
  network_id = data.g42vbcloud_vpc_subnet.mynet.id
  fixed_ip = data.g42vbcloud_vpc_subnet.mynet.gateway_ip
}
`)
}
