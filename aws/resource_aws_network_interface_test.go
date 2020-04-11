package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_network_interface", &resource.Sweeper{
		Name: "aws_network_interface",
		F:    testSweepEc2NetworkInterfaces,
		Dependencies: []string{
			"aws_instance",
		},
	})
}

func testSweepEc2NetworkInterfaces(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeNetworkInterfacesPages(&ec2.DescribeNetworkInterfacesInput{}, func(page *ec2.DescribeNetworkInterfacesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, networkInterface := range page.NetworkInterfaces {
			id := aws.StringValue(networkInterface.NetworkInterfaceId)

			if aws.StringValue(networkInterface.Status) != ec2.NetworkInterfaceStatusAvailable {
				log.Printf("[INFO] Skipping EC2 Network Interface in unavailable (%s) status: %s", aws.StringValue(networkInterface.Status), id)
				continue
			}

			input := &ec2.DeleteNetworkInterfaceInput{
				NetworkInterfaceId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Network Interface: %s", id)
			_, err := conn.DeleteNetworkInterface(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 Network Interface (%s): %s", id, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Network Interface sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EC2 Network Interfaces: %s", err)
	}

	return nil
}

func TestAccAWSENI_basic(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test_interface"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
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

func TestAccAWSENI_ipv6(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIIPV6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
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

func TestAccAWSENI_disappears(t *testing.T) {
	var networkInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface),
					testAccCheckAWSENIDisappears(&networkInterface),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSENI_updatedDescription(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigUpdatedDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated ENI Description"),
				),
			},
		},
	})
}

func TestAccAWSENI_attached(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigWithAttachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributesWithAttachment(&conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test_interface"),
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

func TestAccAWSENI_ignoreExternalAttachment(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigExternalAttachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					testAccCheckAWSENIMakeExternalAttachment("aws_instance.foo", &conf),
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

func TestAccAWSENI_sourceDestCheck(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigWithSourceDestCheck,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
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

func TestAccAWSENI_computedIPs(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigWithNoPrivateIPs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
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

func TestAccAWSENI_PrivateIpsCount(t *testing.T) {
	var networkInterface1, networkInterface2, networkInterface3, networkInterface4 ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigPrivateIpsCount(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface1),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface2),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface3),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface4),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
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

func testAccCheckAWSENIExists(n string, res *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ENI ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		describe_network_interfaces_request := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(rs.Primary.ID)},
		}
		describeResp, err := conn.DescribeNetworkInterfaces(describe_network_interfaces_request)

		if err != nil {
			return err
		}

		if len(describeResp.NetworkInterfaces) != 1 ||
			*describeResp.NetworkInterfaces[0].NetworkInterfaceId != rs.Primary.ID {
			return fmt.Errorf("ENI not found")
		}

		*res = *describeResp.NetworkInterfaces[0]

		return nil
	}
}

func testAccCheckAWSENIAttributes(conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if conf.Attachment != nil {
			return fmt.Errorf("expected attachment to be nil")
		}

		if *conf.AvailabilityZone != "us-west-2a" {
			return fmt.Errorf("expected availability_zone to be us-west-2a, but was %s", *conf.AvailabilityZone)
		}

		if len(conf.Groups) != 1 && *conf.Groups[0].GroupName != "foo" {
			return fmt.Errorf("expected security group to be foo, but was %#v", conf.Groups)
		}

		if *conf.PrivateIpAddress != "172.16.10.100" {
			return fmt.Errorf("expected private ip to be 172.16.10.100, but was %s", *conf.PrivateIpAddress)
		}

		if *conf.PrivateDnsName != "ip-172-16-10-100.us-west-2.compute.internal" {
			return fmt.Errorf("expected private dns name to be ip-172-16-10-100.us-west-2.compute.internal, but was %s", *conf.PrivateDnsName)
		}

		if len(*conf.MacAddress) == 0 {
			return fmt.Errorf("expected mac_address to be set")
		}

		if !*conf.SourceDestCheck {
			return fmt.Errorf("expected source_dest_check to be true, but was %t", *conf.SourceDestCheck)
		}

		if len(conf.TagSet) == 0 {
			return fmt.Errorf("expected tags")
		}

		return nil
	}
}

func testAccCheckAWSENIAttributesWithAttachment(conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if conf.Attachment == nil {
			return fmt.Errorf("expected attachment to be set, but was nil")
		}

		if *conf.Attachment.DeviceIndex != 1 {
			return fmt.Errorf("expected attachment device index to be 1, but was %d", *conf.Attachment.DeviceIndex)
		}

		if *conf.AvailabilityZone != "us-west-2a" {
			return fmt.Errorf("expected availability_zone to be us-west-2a, but was %s", *conf.AvailabilityZone)
		}

		if len(conf.Groups) != 1 && *conf.Groups[0].GroupName != "foo" {
			return fmt.Errorf("expected security group to be foo, but was %#v", conf.Groups)
		}

		if *conf.PrivateIpAddress != "172.16.10.100" {
			return fmt.Errorf("expected private ip to be 172.16.10.100, but was %s", *conf.PrivateIpAddress)
		}

		if *conf.PrivateDnsName != "ip-172-16-10-100.us-west-2.compute.internal" {
			return fmt.Errorf("expected private dns name to be ip-172-16-10-100.us-west-2.compute.internal, but was %s", *conf.PrivateDnsName)
		}

		return nil
	}
}

func testAccCheckAWSENIDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network_interface" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		describe_network_interfaces_request := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(rs.Primary.ID)},
		}
		_, err := conn.DescribeNetworkInterfaces(describe_network_interfaces_request)

		if err != nil {
			if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
				return nil
			}

			return err
		}
	}

	return nil
}

func testAccCheckAWSENIDisappears(networkInterface *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: networkInterface.NetworkInterfaceId,
		}
		_, err := conn.DeleteNetworkInterface(input)

		return err
	}
}

func testAccCheckAWSENIMakeExternalAttachment(n string, conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("Not found: %s", n)
		}
		attach_request := &ec2.AttachNetworkInterfaceInput{
			DeviceIndex:        aws.Int64(1),
			InstanceId:         aws.String(rs.Primary.ID),
			NetworkInterfaceId: conf.NetworkInterfaceId,
		}
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, attach_err := conn.AttachNetworkInterface(attach_request)
		if attach_err != nil {
			return fmt.Errorf("Error attaching ENI: %s", attach_err)
		}
		return nil
	}
}

const testAccAWSENIConfig = `
resource "aws_vpc" "test" {
	cidr_block = "172.16.0.0/16"
	enable_dns_hostnames = true
	tags = {
		Name = "terraform-testacc-network-interface"
	}
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-network-interface"
    }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test"
  name = "tf-acc-network-interface"

        egress {
                from_port = 0
                to_port = 0
                protocol = "tcp"
                cidr_blocks = ["10.0.0.0/16"]
        }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.test.id}"]
    description = "Managed by Terraform"
  tags = {
        Name = "test_interface"
    }
}
`

const testAccAWSENIIPV6Config = `
resource "aws_vpc" "test" {
  cidr_block                       = "172.16.0.0/16"
  assign_generated_ipv6_cidr_block = true
  enable_dns_hostnames             = true

  tags = {
  	Name = "terraform-testacc-network-interface-ipv6"
  }
}

resource "aws_subnet" "test" {
  vpc_id                          = "${aws_vpc.test.id}"
  cidr_block                      = "172.16.10.0/24"
  ipv6_cidr_block                 = "${cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 16)}"
  availability_zone               = "us-west-2a"

  tags = {
    Name = "tf-acc-network-interface-ipv6"
  }
}

resource "aws_security_group" "test" {
  vpc_id      = "${aws_vpc.test.id}"
  description = "test"
  name        = "tf-acc-network-interface-ipv6"

  tags = {
    Name = "test-interface-ipv6"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = "${aws_subnet.test.id}"
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = ["${cidrhost(aws_subnet.test.ipv6_cidr_block, 4)}"]
  security_groups = ["${aws_security_group.test.id}"]
  description     = "Managed by Terraform"

  tags = {
    Name = "test-interface-ipv6"
  }
}
`

const testAccAWSENIConfigUpdatedDescription = `
resource "aws_vpc" "test" {
	cidr_block = "172.16.0.0/16"
	enable_dns_hostnames = true
	tags = {
		Name = "terraform-testacc-network-interface-update-desc"
	}
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-network-interface-update-desc"
    }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test"
  name = "terraform-testacc-network-interface-update-desc"

        egress {
                from_port = 0
                to_port = 0
                protocol = "tcp"
                cidr_blocks = ["10.0.0.0/16"]
        }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.test.id}"]
    description = "Updated ENI Description"
  tags = {
        Name = "test_interface"
    }
}
`

const testAccAWSENIConfigWithSourceDestCheck = `
resource "aws_vpc" "test" {
	cidr_block = "172.16.0.0/16"
	enable_dns_hostnames = true
	tags = {
		Name = "terraform-testacc-network-interface-w-source-dest-check"
	}
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-network-interface-w-source-dest-check"
    }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
        source_dest_check = false
    private_ips = ["172.16.10.100"]
}
`

const testAccAWSENIConfigWithNoPrivateIPs = `
resource "aws_vpc" "test" {
	cidr_block = "172.16.0.0/16"
	enable_dns_hostnames = true
	tags = {
		Name = "terraform-testacc-network-interface-w-no-private-ips"
	}
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-network-interface-w-no-private-ips"
    }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
        source_dest_check = false
}
`

const testAccAWSENIConfigWithAttachment = `
resource "aws_vpc" "test" {
	cidr_block = "172.16.0.0/16"
	enable_dns_hostnames = true
  tags = {
        Name = "terraform-testacc-network-interface-w-attachment"
    }
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
            Name = "tf-acc-network-interface-w-attachment-test"
        }
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.11.0/24"
    availability_zone = "us-west-2a"
  tags = {
            Name = "tf-acc-network-interface-w-attachment-test"
        }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test"
  name = "tf-acc-network-interface-w-attachment-test"
}

resource "aws_instance" "test" {
    ami = "ami-c5eabbf5"
    instance_type = "t2.micro"
    subnet_id = "${aws_subnet.test.id}"
    associate_public_ip_address = false
    private_ip = "172.16.11.50"
  tags = {
        Name = "test-tf-eni-test"
    }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.test.id}"]
    attachment {
        instance = "${aws_instance.test.id}"
        device_index = 1
    }
  tags = {
        Name = "test_interface"
    }
}
`

const testAccAWSENIConfigExternalAttachment = `
resource "aws_vpc" "test" {
	cidr_block = "172.16.0.0/16"
	enable_dns_hostnames = true
  tags = {
        Name = "terraform-testacc-network-interface-external-attachment"
    }
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-network-interface-external-attachment-test"
    }
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.11.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-network-interface-external-attachment-test"
    }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test"
  name = "tf-acc-network-interface-external-attachment-test"
}

resource "aws_instance" "test" {
    ami = "ami-c5eabbf5"
    instance_type = "t2.micro"
    subnet_id = "${aws_subnet.test.id}"
    associate_public_ip_address = false
    private_ip = "172.16.11.50"
  tags = {
        Name = "tf-eni-test"
    }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.test.id}"]
  tags = {
        Name = "test_interface"
    }
}
`

func testAccAWSENIConfigPrivateIpsCount(privateIpsCount int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-network-interface-private-ips-count"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-network-interface-private-ips-count"
  }
}

resource "aws_network_interface" "test" {
  private_ips_count = %[1]d
  subnet_id         = "${aws_subnet.test.id}"
}
`, privateIpsCount)
}
