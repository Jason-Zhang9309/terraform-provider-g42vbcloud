package g42vbcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk/openstack/evs/v3/volumes"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
)

func TestAccEvsStorageV3Volume_basic(t *testing.T) {
	var volume volumes.Volume

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "g42vbcloud_evs_volume.test"
	rNameUpdate := rName + "-updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEvsStorageV3VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEvsStorageV3Volume_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEvsStorageV3VolumeExists(resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccEvsStorageV3Volume_basic(rNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEvsStorageV3VolumeExists(resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdate),
				),
			},
		},
	})
}

func TestAccEvsStorageV3Volume_image(t *testing.T) {
	var volume volumes.Volume

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "g42vbcloud_evs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEvsStorageV3VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEvsStorageV3Volume_image(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEvsStorageV3VolumeExists(resourceName, &volume),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func testAccCheckEvsStorageV3VolumeDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*config.Config)
	blockStorageClient, err := config.BlockStorageV3Client(G42VB_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating G42VBCloud evs storage client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "g42vbcloud_evs_volume_v3" {
			continue
		}

		_, err := volumes.Get(blockStorageClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Volume still exists")
		}
	}

	return nil
}

func testAccCheckEvsStorageV3VolumeExists(n string, volume *volumes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*config.Config)
		blockStorageClient, err := config.BlockStorageV3Client(G42VB_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating G42VBCloud evs storage client: %s", err)
		}

		found, err := volumes.Get(blockStorageClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Volume not found")
		}

		*volume = *found

		return nil
	}
}

func testAccCheckEvsStorageV3VolumeTags(n string, k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*config.Config)
		blockStorageClient, err := config.BlockStorageV3Client(G42VB_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating G42VBCloud block storage client: %s", err)
		}

		found, err := volumes.Get(blockStorageClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Volume not found")
		}

		if found.Tags == nil {
			return fmt.Errorf("No Tags")
		}

		for key, value := range found.Tags {
			if k != key {
				continue
			}

			if v == value {
				return nil
			}
			return fmt.Errorf("Bad value for %s: %s", k, value)
		}
		return fmt.Errorf("Tag not found: %s", k)
	}
}

func testAccEvsStorageV3Volume_basic(rName string) string {
	return fmt.Sprintf(`
data "g42vbcloud_availability_zones" "test" {}

resource "g42vbcloud_evs_volume" "test" {
  name              = "%s"
  description       = "test volume"
  availability_zone = data.g42vbcloud_availability_zones.test.names[0]
  volume_type       = "SAS"
  size              = 12
}
`, rName)
}

func testAccEvsStorageV3Volume_image(rName string) string {
	return fmt.Sprintf(`
data "g42vbcloud_availability_zones" "test" {}

data "g42vbcloud_images_image" "test" {
  name        = "Ubuntu 18.04 server 64bit"
  most_recent = true
}

resource "g42vbcloud_evs_volume" "test" {
  name              = "%s"
  availability_zone = data.g42vbcloud_availability_zones.test.names[0]
  volume_type       = "SAS"
  size              = 40
  image_id          = data.g42vbcloud_images_image.test.id
}
`, rName)
}
