package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDeviceFarmProject_basic(t *testing.T) {
	var proj devicefarm.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSDeviceFarmProject_disappears(t *testing.T) {
	var proj devicefarm.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					testAccCheckDeviceFarmProjectDisappears(&proj),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFarmProjectExists(n string, v *devicefarm.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).devicefarmconn
		resp, err := conn.GetProject(
			&devicefarm.GetProjectInput{Arn: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if resp.Project == nil {
			return fmt.Errorf("DeviceFarmProject not found")
		}

		*v = *resp.Project

		return nil
	}
}

func testAccCheckDeviceFarmProjectDisappears(v *devicefarm.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).devicefarmconn
		_, err := conn.DeleteProject(
			&devicefarm.DeleteProjectInput{Arn: v.Arn})

		return err
	}
}

func testAccCheckDeviceFarmProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).devicefarmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_devicefarm_project" {
			continue
		}

		// Try to find the resource
		resp, err := conn.GetProject(
			&devicefarm.GetProjectInput{Arn: aws.String(rs.Primary.ID)})
		if err == nil {
			if resp.Project != nil {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		if isAWSErr(err, devicefarm.ErrCodeNotFoundException, "") {
			return nil
		}
	}

	return nil
}

func testAccDeviceFarmProjectConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_project" "test" {
  name = %[1]q
}
`, rName)
}
