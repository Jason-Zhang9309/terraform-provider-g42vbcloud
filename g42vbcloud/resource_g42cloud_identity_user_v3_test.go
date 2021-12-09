package g42vbcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk/openstack/identity/v3/users"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
)

func TestAccIdentityV3User_basic(t *testing.T) {
	var user users.User
	var userName = fmt.Sprintf("ACCPTTEST-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAdminOnly(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIdentityV3UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3User_basic(userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3UserExists("g42vbcloud_identity_user_v3.user_1", &user),
					resource.TestCheckResourceAttrPtr(
						"g42vbcloud_identity_user_v3.user_1", "name", &user.Name),
					resource.TestCheckResourceAttr(
						"g42vbcloud_identity_user_v3.user_1", "enabled", "true"),
				),
			},
			{
				Config: testAccIdentityV3User_update(userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3UserExists("g42vbcloud_identity_user_v3.user_1", &user),
					resource.TestCheckResourceAttrPtr(
						"g42vbcloud_identity_user_v3.user_1", "name", &user.Name),
					resource.TestCheckResourceAttr(
						"g42vbcloud_identity_user_v3.user_1", "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckIdentityV3UserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*config.Config)
	identityClient, err := config.IdentityV3Client(G42VB_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating G42VBCloud identity client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "g42vbcloud_identity_user_v3" {
			continue
		}

		_, err := users.Get(identityClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("User still exists")
		}
	}

	return nil
}

func testAccCheckIdentityV3UserExists(n string, user *users.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*config.Config)
		identityClient, err := config.IdentityV3Client(G42VB_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating G42VBCloud identity client: %s", err)
		}

		found, err := users.Get(identityClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("User not found")
		}

		*user = *found

		return nil
	}
}

func testAccIdentityV3User_basic(userName string) string {
	return fmt.Sprintf(`
    resource "g42vbcloud_identity_user_v3" "user_1" {
      name = "%s"
      password = "password123@!"
      enabled = true
      description = "tested by terraform"
    }  
  `, userName)
}

func testAccIdentityV3User_update(userName string) string {
	return fmt.Sprintf(`
    resource "g42vbcloud_identity_user_v3" "user_1" {
      name = "%s"
      enabled = false
      password = "password123@!"
      description = "tested by terraform"
    }
  `, userName)
}
