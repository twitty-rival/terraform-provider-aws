package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCloudTrail(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Trail": {
			"basic":                      testAccAWSCloudTrail_basic,
			"cloudwatch":                 testAccAWSCloudTrail_cloudwatch,
			"enableLogging":              testAccAWSCloudTrail_enable_logging,
			"includeGlobalServiceEvents": testAccAWSCloudTrail_include_global_service_events,
			"isMultiRegion":              testAccAWSCloudTrail_is_multi_region,
			"isOrganization":             testAccAWSCloudTrail_is_organization,
			"logValidation":              testAccAWSCloudTrail_logValidation,
			"kmsKey":                     testAccAWSCloudTrail_kmsKey,
			"tags":                       testAccAWSCloudTrail_tags,
			"eventSelector":              testAccAWSCloudTrail_event_selector,
			"eventSelectorWExclude":      testAccAWSCloudTrail_event_selector_exclude,
			"disappears":                 testAccAWSCloudTrail_disappears,
			"sns":                        testAccAWSCloudTrail_sns,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAWSCloudTrail_basic(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "cloudtrail", fmt.Sprintf("trail/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_sns(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-sns")
	resourceName := "aws_cloudtrail.test"
	topicResourceName := "aws_sns_topic.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigSNS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_name", topicResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", topicResourceName, "arn"),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_cloudwatch(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-cw")
	resourceName := "aws_cloudtrail.test"
	cwResourceName := "aws_cloudwatch_log_group.test"
	cwChangeResourceName := "aws_cloudwatch_log_group.second"
	roleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigCloudWatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_watch_logs_group_arn", cwResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_watch_logs_role_arn", roleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigCloudWatchModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_watch_logs_group_arn", cwChangeResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_watch_logs_role_arn", roleResourceName, "arn"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_enable_logging(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-log")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					// AWS will create the trail with logging turned off.
					// Test that "enable_logging" default works.
					testAccCheckCloudTrailLoggingEnabled(resourceName, true),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, false),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, true),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_tags(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
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
				Config: testAccAWSCloudTrailConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudTrailConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_is_multi_region(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-multi")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				Config: testAccAWSCloudTrailConfigMultiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_is_organization(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-org")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigOrganization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_logValidation(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-log-val")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_logValidation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, true, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig_logValidationModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals(resourceName, "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_kmsKey(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-kms")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					testAccMatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile("key/.+")),
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

func testAccAWSCloudTrail_include_global_service_events(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-events")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_include_global_service_events(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "false"),
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

func testAccAWSCloudTrail_event_selector(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-event")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_eventSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s/testbar", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s/baz", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig_eventSelectorModified(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s/foobar", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s/baz", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.0", "s3", fmt.Sprintf("%s/tf1", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.1", "s3", fmt.Sprintf("%s/tf2", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.type", "AWS::Lambda::Function"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.values.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "event_selector.1.data_resource.1.values.0", "lambda", regexp.MustCompile(`function:.+`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.read_write_type", "All"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "0"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_event_selector_exclude(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-exclude")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_eventSelectorExclude(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s/foobar", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s/baz", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.0", "kms.amazonaws.com"),
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

func testAccAWSCloudTrail_disappears(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-cloudtrail-test-disappears")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailDisappears(&trail),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudTrailExists(n string, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeTrails(&params)
		if err != nil {
			return err
		}
		if len(resp.TrailList) == 0 {
			return fmt.Errorf("Trail not found")
		}
		*trail = *resp.TrailList[0]

		return nil
	}
}

func testAccCheckCloudTrailDisappears(trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn

		input := &cloudtrail.DeleteTrailInput{Name: trail.Name}
		_, err := conn.DeleteTrail(input)

		return err
	}
}

func testAccCheckCloudTrailLoggingEnabled(n string, desired bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		params := cloudtrail.GetTrailStatusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTrailStatus(&params)

		if err != nil {
			return err
		}
		if *resp.IsLogging != desired {
			return fmt.Errorf("Expected logging status %t, given %t", desired, *resp.IsLogging)
		}

		return nil
	}
}

func testAccCheckCloudTrailLogValidationEnabled(n string, desired bool, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if trail.LogFileValidationEnabled == nil {
			return fmt.Errorf("No LogFileValidationEnabled attribute present in trail: %s", trail)
		}

		if *trail.LogFileValidationEnabled != desired {
			return fmt.Errorf("Expected log validation status %t, given %t", desired,
				*trail.LogFileValidationEnabled)
		}

		// local state comparison
		enabled, ok := rs.Primary.Attributes["enable_log_file_validation"]
		if !ok {
			return fmt.Errorf("No enable_log_file_validation attribute defined for %s, expected %t",
				n, desired)
		}
		desiredInString := fmt.Sprintf("%t", desired)
		if enabled != desiredInString {
			return fmt.Errorf("Expected log validation status %s, saved %s", desiredInString, enabled)
		}

		return nil
	}
}

func testAccCheckCloudTrailKmsKeyIdEquals(n string, desired string, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if desired != "" && trail.KmsKeyId == nil {
			return fmt.Errorf("No KmsKeyId attribute present in trail: %s, expected %s",
				trail, desired)
		}

		// work around string pointer
		var kmsKeyIdInString string
		if trail.KmsKeyId == nil {
			kmsKeyIdInString = ""
		} else {
			kmsKeyIdInString = *trail.KmsKeyId
		}

		if kmsKeyIdInString != desired {
			return fmt.Errorf("Expected KMS Key ID %q to equal %q",
				*trail.KmsKeyId, desired)
		}

		kmsKeyId, ok := rs.Primary.Attributes["kms_key_id"]
		if desired != "" && !ok {
			return fmt.Errorf("No kms_key_id attribute defined for %s", n)
		}
		if kmsKeyId != desired {
			return fmt.Errorf("Expected KMS Key ID %q, saved %q", desired, kmsKeyId)
		}

		return nil
	}
}

func testAccCheckAWSCloudTrailDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudtrail" {
			continue
		}

		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeTrails(&params)

		if err == nil {
			if len(resp.TrailList) != 0 &&
				*resp.TrailList[0].Name == rs.Primary.ID {
				return fmt.Errorf("CloudTrail still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSCloudTrailConfigS3Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::%[1]s"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::%[1]s/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, rName)
}

func testAccAWSCloudTrailConfig(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"
}
`, rName)
}

func testAccAWSCloudTrailConfigModified(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = "${aws_s3_bucket.test.id}"
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
  enable_logging                = false
}
`, rName)
}

func testAccAWSCloudTrailConfigSNS(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"
  sns_topic_name = "${aws_sns_topic.test.name}"
}

resource "aws_sns_topic" "test" {
  name   = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn = "${aws_sns_topic.test.arn}"
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AWSCloudTrailSNSPolicy20131101",
            "Effect": "Allow",
            "Principal": {
                "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "SNS:Publish",
            "Resource": "${aws_sns_topic.test.arn}"
        }
    ]
}
POLICY
}
`, rName)
}

func testAccAWSCloudTrailConfigCloudWatch(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test.arn}"
  cloud_watch_logs_role_arn  = "${aws_iam_role.test.arn}"
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = "${aws_iam_role.test.id}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailCreateLogStream",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "${aws_cloudwatch_log_group.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccAWSCloudTrailConfigCloudWatchModified(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.second.arn}"
  cloud_watch_logs_role_arn  = "${aws_iam_role.test.arn}"
}


resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "second" {
  name = "%[1]s-2"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = "${aws_iam_role.test.id}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailCreateLogStream",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "${aws_cloudwatch_log_group.second.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccAWSCloudTrailConfigMultiRegion(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                  = %[1]q
  s3_bucket_name        = "${aws_s3_bucket.test.id}"
  is_multi_region_trail = true
}
`, rName)
}

func testAccAWSCloudTrailConfigOrganization(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["cloudtrail.amazonaws.com"]
}

resource "aws_cloudtrail" "test" {
  is_organization_trail = true
  name                  = %q
  s3_bucket_name        = "${aws_s3_bucket.test.id}"
}
`, rName)
}

func testAccAWSCloudTrailConfig_logValidation(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %q
  s3_bucket_name                = "${aws_s3_bucket.test.id}"
  is_multi_region_trail         = true
  include_global_service_events = true
  enable_log_file_validation    = true
}
`, rName)
}

func testAccAWSCloudTrailConfig_logValidationModified(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = "${aws_s3_bucket.test.id}"
  include_global_service_events = true
}
`, rName)
}

func testAccAWSCloudTrailConfig_kmsKey(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = "${aws_s3_bucket.test.id}"
  include_global_service_events = true
  kms_key_id                    = "${aws_kms_key.test.arn}"
}
`, rName)
}

func testAccAWSCloudTrailConfig_include_global_service_events(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %q
  s3_bucket_name                = "${aws_s3_bucket.test.id}"
  include_global_service_events = false
}
`, rName)
}

func testAccAWSCloudTrailConfig_eventSelector(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test.arn}/testbar",
        "${aws_s3_bucket.test.arn}/baz",
      ]
    }
  }
}
`, rName)
}

func testAccAWSCloudTrailConfig_eventSelectorExclude(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  event_selector {
    read_write_type                  = "ReadOnly"
    include_management_events        = true
    exclude_management_event_sources = ["kms.amazonaws.com"]


    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test.arn}/foobar",
        "${aws_s3_bucket.test.arn}/baz",
      ]
    }
  }
}

`, rName)
}

func testAccAWSCloudTrailConfig_eventSelectorModified(rName string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = true

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test.arn}/foobar",
        "${aws_s3_bucket.test.arn}/baz",
      ]
    }
  }

  event_selector {
    read_write_type           = "All"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test.arn}/tf1",
        "${aws_s3_bucket.test.arn}/tf2",
      ]
    }

    data_resource {
      type = "AWS::Lambda::Function"

      values = [
        "${aws_lambda_function.test.arn}",
      ]
    }
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.test.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccAWSCloudTrailConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSCloudTrailConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSCloudTrailConfigS3Base(rName) + fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
