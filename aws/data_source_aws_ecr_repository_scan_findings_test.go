package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEcrRepositoryScanFindingsDataSource_basic(t *testing.T) {
	registry, repo, tag := "137112412989", "amazonlinux", "latest"
	dataSourceName := "data.aws_ecr_repository_scan_findings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrRepositoryScanFindingsDataSourceConfig(registry, repo, tag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "scan_findings"),
				),
			},
		},
	})
}

func testAccCheckAwsEcrRepositoryScanFindingsDataSourceConfig(reg, repo, tag string) string {
	return fmt.Sprintf(`
data "aws_ecr_repository_scan_findings" "test" {
  registry_id     = "%s"
  repository_name = "%s"
  image_tag       = "%s"
}
`, reg, repo, tag)
}
