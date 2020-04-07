package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSLBCookieStickinessPolicy_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-lb-%s", acctest.RandString(5))
	resourceName := "aws_lb_cookie_stickiness_policy.test"
	lbResourceName := "aws_elb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(lbResourceName, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLBCookieStickinessPolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(lbResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckLBCookieStickinessPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_cookie_stickiness_policy" {
			continue
		}

		lbName, _, policyName := resourceAwsLBCookieStickinessPolicyParseId(rs.Primary.ID)
		out, err := conn.DescribeLoadBalancerPolicies(
			&elb.DescribeLoadBalancerPoliciesInput{
				LoadBalancerName: aws.String(lbName),
				PolicyNames:      []*string{aws.String(policyName)},
			})
		if err != nil {
			if isAWSErr(err, elb.ErrCodePolicyNotFoundException, "") ||
				isAWSErr(err, "LoadBalancerNotFound", "") {
				continue
			}
			return err
		}

		if len(out.PolicyDescriptions) > 0 {
			return fmt.Errorf("Policy still exists")
		}
	}

	return nil
}

func testAccCheckLBCookieStickinessPolicy(elbResource string, policyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[elbResource]
		if !ok {
			return fmt.Errorf("Not found: %s", elbResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[policyResource]
		if !ok {
			return fmt.Errorf("Not found: %s", policyResource)
		}

		conn := testAccProvider.Meta().(*AWSClient).elbconn
		elbName, _, policyName := resourceAwsLBCookieStickinessPolicyParseId(policy.Primary.ID)
		_, err := conn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(elbName),
			PolicyNames:      []*string{aws.String(policyName)},
		})

		return err
	}
}

func TestAccAWSLBCookieStickinessPolicy_drift(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", acctest.RandString(5))
	resourceName := "aws_lb_cookie_stickiness_policy.test"
	lbResourceName := "aws_elb.test"

	// We only want to remove the reference to the policy from the listner,
	// beacause that's all that can be done via the console.
	removePolicy := func() {
		conn := testAccProvider.Meta().(*AWSClient).elbconn

		setLoadBalancerOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
			LoadBalancerName: aws.String(lbName),
			LoadBalancerPort: aws.Int64(80),
			PolicyNames:      []*string{},
		}

		if _, err := conn.SetLoadBalancerPoliciesOfListener(setLoadBalancerOpts); err != nil {
			t.Fatalf("Error removing LBCookieStickinessPolicy: %s", err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(lbResourceName, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: removePolicy,
				Config:    testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(lbResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccAWSLBCookieStickinessPolicy_missingLB(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", acctest.RandString(5))
	resourceName := "aws_lb_cookie_stickiness_policy.test"
	lbResourceName := "aws_elb.test"

	// check that we can destroy the policy if the LB is missing
	removeLB := func() {
		conn := testAccProvider.Meta().(*AWSClient).elbconn
		deleteElbOpts := elb.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(lbName),
		}
		if _, err := conn.DeleteLoadBalancer(&deleteElbOpts); err != nil {
			t.Fatalf("Error deleting ELB: %s", err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(lbResourceName, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: removeLB,
				Config:    testAccLBCookieStickinessPolicyConfigDestroy(lbName),
			},
		},
	})
}

func testAccLBCookieStickinessPolicyConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_lb_cookie_stickiness_policy" "test" {
  name          = %[1]q
  load_balancer = "${aws_elb.test.id}"
  lb_port       = 80
}
`, rName)
}

// Sets the cookie_expiration_period to 300s.
func testAccLBCookieStickinessPolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_lb_cookie_stickiness_policy" "test" {
  name                     = %[1]q
  load_balancer            = "${aws_elb.test.id}"
  lb_port                  = 80
  cookie_expiration_period = 300
}
`, rName)
}

// attempt to destroy the policy, but we'll delete the LB in the PreConfig
func testAccLBCookieStickinessPolicyConfigDestroy(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`, rName)
}
