package aws

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsEcrRepositoryScanFindings() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcrRepositoryScanFindingsRead,

		Schema: map[string]*schema.Schema{
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"image_tag": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scan_findings": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEcrRepositoryScanFindingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	imgId := ecr.ImageIdentifier{
		ImageTag: aws.String(d.Get("image_tag").(string)),
	}

	params := &ecr.DescribeImageScanFindingsInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
		ImageId:        &imgId,
		RegistryId:     aws.String(d.Get("registry_id").(string)),
	}
	log.Printf("[DEBUG] Reading ECR repository scan findings: %s", params)
	out, err := conn.DescribeImageScanFindings(params)
	if err != nil {
		if isAWSErr(err, ecr.ErrCodeScanNotFoundException, "") {
			log.Printf("[WARN] ECR Repository scan findings %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading ECR repository scan findings: %s", err)
	}

	scanFindings := out.ImageScanFindings
	scanFindingsJson, err := json.Marshal(scanFindings.Findings)
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	jsonString := string(scanFindingsJson)
	d.Set("scan_findings", jsonString)

	return nil
}
