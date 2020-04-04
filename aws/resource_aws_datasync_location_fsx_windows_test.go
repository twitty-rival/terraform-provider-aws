package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_fsx_windows", &resource.Sweeper{
		Name: "aws_datasync_location_fsx_windows",
		F:    testSweepDataSyncLocationFsxWindows,
	})
}

func testSweepDataSyncLocationFsxWindows(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location FSX Windows sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location FSX Windows: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location FSX Windowss to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "FsxWindows://") {
				log.Printf("[INFO] Skipping DataSync Location FSX Windows: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location FSX Windows: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location FSX Windows (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationFsxWindows_basic(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^fsxw://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_disappears(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					testAccCheckAWSDataSyncLocationFsxWindowsDisappears(&locationFsxWindows1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_Subdirectory(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigSubdirectory("/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_Tags(t *testing.T) {
	var locationFsxWindows1, locationFsxWindows2, locationFsxWindows3 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows2),
					testAccCheckAWSDataSyncLocationFsxWindowsNotRecreated(&locationFsxWindows1, &locationFsxWindows2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows3),
					testAccCheckAWSDataSyncLocationFsxWindowsNotRecreated(&locationFsxWindows2, &locationFsxWindows3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncLocationFsxWindowsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_fsx_windows" {
			continue
		}

		input := &datasync.DescribeLocationFsxWindowsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationFsxWindows(input)

		if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName string, locationFsxWindows *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn
		input := &datasync.DescribeLocationFsxWindowsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationFsxWindows(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationFsxWindows = *output

		return nil
	}
}

func testAccCheckAWSDataSyncLocationFsxWindowsDisappears(location *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckAWSDataSyncLocationFsxWindowsNotRecreated(i, j *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("DataSync Location FSX Windows was recreated")
		}

		return nil
	}
}

func testAccAWSDataSyncLocationFsxWindowsConfig() string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  fsx_filesystem_arn  = "${aws_fsx_windows_file_system.test.arn}"
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = ["${aws_security_group.test1.arn}"]
}
`)
}

func testAccAWSDataSyncLocationFsxWindowsConfigSubdirectory(subdirectory string) string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  fsx_filesystem_arn  = "${aws_fsx_windows_file_system.test.arn}"
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = ["${aws_security_group.test1.arn}"]
  subdirectory        = %q
}
`, subdirectory)
}

func testAccAWSDataSyncLocationFsxWindowsConfigTags1(key1, value1 string) string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  fsx_filesystem_arn  = "${aws_fsx_windows_file_system.test.arn}"
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = ["${aws_security_group.test1.arn}"]

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1)
}

func testAccAWSDataSyncLocationFsxWindowsConfigTags2(key1, value1, key2, value2 string) string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  fsx_filesystem_arn  = "${aws_fsx_windows_file_system.test.arn}"
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = ["${aws_security_group.test1.arn}"]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2)
}
