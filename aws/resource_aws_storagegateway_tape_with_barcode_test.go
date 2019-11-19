package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSStorageGatewayTapeWithBarcode_basic(t *testing.T) {
	var tapeWithBarcode storagegateway.Tape
	rName := acctest.RandomWithPrefix("tf-acc-test")
	barcode := "TEST12345"
	resourceName := "aws_storagegateway_tape_with_barcode.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapeWithBarcodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfig_Required(rName, barcode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					testAccMatchResourceAttrRegionalARN(resourceName, "gateway_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tape_barcode", barcode),
					resource.TestCheckResourceAttr(resourceName, "tape_size_in_bytes", "5368709120"),
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

func TestAccAWSStorageGatewayTapeWithBarcode_tags(t *testing.T) {
	var tapeWithBarcode storagegateway.Tape
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_with_barcode.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapeWithBarcodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayTapeWithBarcode_KMSEncrypted(t *testing.T) {
	var tapeWithBarcode storagegateway.Tape
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_with_barcode.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapeWithBarcodeDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSEncrypted(rName, true),
				ExpectError: regexp.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
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

func TestAccAWSStorageGatewayTapeWithBarcode_KMSKeyArn(t *testing.T) {
	var tapeWithBarcode storagegateway.Tape
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_with_barcode.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapeWithBarcodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSKeyArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestMatchResourceAttr(resourceName, "kms_key_arn", regexp.MustCompile(`^arn:`)),
				),
			},
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSKeyArn_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestMatchResourceAttr(resourceName, "kms_key_arn", regexp.MustCompile(`^arn:`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName, &tapeWithBarcode),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSStorageGatewayTapeWithBarcodeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_tape_with_barcode" {
			continue
		}

		input := &storagegateway.DescribeTapesInput{
			TapeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeTapes(input)

		if err != nil {
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				continue
			}
			return err
		}

		if output != nil && len(output.Tapes) > 0 && output.Tapes[0] != nil {
			return fmt.Errorf("Storage Gateway Tape With Barcode %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAWSStorageGatewayTapeWithBarcodeExists(resourceName string, tapeWithBarcode *storagegateway.Tape) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeTapesInput{
			TapeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeTapes(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Tapes) == 0 || output.Tapes[0] == nil {
			return fmt.Errorf("Storage Gateway Tape With Barcode %q does not exist", rs.Primary.ID)
		}

		*tapeWithBarcode = *output.Tapes[0]

		return nil
	}
}

func testAccAWSStorageGatewayTapeWithBarcodeConfig_Required(rName, barcode string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName) + fmt.Sprintf(`
resource "aws_storagegateway_tape_with_barcode" "test" {
  tape_barcode       = "%s"
  tape_size_in_bytes = "5368709120"
  gateway_arn        = "${aws_storagegateway_gateway.test.arn}"
}
`, barcode)
}

func testAccAWSStorageGatewayTapeWithBarcodeConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName) + fmt.Sprintf(`
resource "aws_storagegateway_tape_with_barcode" "test" {
  gateway_arn  = "${aws_storagegateway_gateway.test.arn}"

  tags = {
	%q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSStorageGatewayTapeWithBarcodeConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName) + fmt.Sprintf(`
resource "aws_storagegateway_tape_with_barcode" "test" {
  gateway_arn  = "${aws_storagegateway_gateway.test.arn}"

  tags = {
	%q = %q
	%q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSEncrypted(rName string, kmsEncrypted bool) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName) + fmt.Sprintf(`
resource "aws_storagegateway_tape_with_barcode" "test" {
  gateway_arn   = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted = %t
}
`, kmsEncrypted)
}

func testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSKeyArn(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_tape_with_barcode" "test" {
  gateway_arn             = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted           = true
  kms_key_arn             = "${aws_kms_key.test.0.arn}"
}
`
}

func testAccAWSStorageGatewayTapeWithBarcodeConfig_KMSKeyArn_Update(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_tape_with_barcode" "test" {
  gateway_arn             = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted           = true
  kms_key_arn             = "${aws_kms_key.test.1.arn}"
}
`
}
