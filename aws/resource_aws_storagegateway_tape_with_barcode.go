package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsStorageGatewayTapeWithBarcode() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsStorageGatewayTapeWithBarcodeCreate,
		Read:   resourceAwsStorageGatewayTapeWithBarcodeRead,
		Update: resourceAwsStorageGatewayTapeWithBarcodeUpdate,
		Delete: resourceAwsStorageGatewayTapeWithBarcodeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"tape_barcode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tape_size_in_bytes": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"tape_used_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "GLACIER",
				ValidateFunc: validation.StringInSlice([]string{
					"GLACIER",
					"DEEP_ARCHIVE",
				}, false),
			},
			"vtl_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsStorageGatewayTapeWithBarcodeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.CreateTapeWithBarcodeInput{
		GatewayARN:      aws.String(d.Get("gateway_arn").(string)),
		KMSEncrypted:    aws.Bool(d.Get("kms_encrypted").(bool)),
		TapeBarcode:     aws.String(d.Get("tape_barcode").(string)),
		TapeSizeInBytes: aws.Int64(int64(d.Get("tape_size_in_bytes").(int))),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().StoragegatewayTags(),
	}

	if v, ok := d.GetOk("pool_id"); ok && v.(string) != "" {
		input.PoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok && v.(string) != "" {
		input.KMSKey = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Storage Gateway Tape With Barcode: %s", input)
	output, err := conn.CreateTapeWithBarcode(input)
	if err != nil {
		return fmt.Errorf("error creating Storage Gateway Tape With Barcode: %s", err)
	}

	d.SetId(aws.StringValue(output.TapeARN))
	d.Set("gateway_arn", d.Get("gateway_arn").(string))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING", "MISSING"},
		Target:     []string{"AVAILABLE"},
		Refresh:    storageGatewayTapeWithBarcodeRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Storage Gateway Tape With Barcode creation: %s", err)
	}

	return resourceAwsStorageGatewayTapeWithBarcodeRead(d, meta)
}

func resourceAwsStorageGatewayTapeWithBarcodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DescribeTapesInput{
		TapeARNs: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway Tape With Barcode: %s", input)
	output, err := conn.DescribeTapes(input)
	if err != nil {
		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified tape was not found.") {
			log.Printf("[WARN] Storage Gateway Tape With Barcode %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Storage Gateway Tape With Barcode: %s", err)
	}

	if output == nil || len(output.Tapes) == 0 || output.Tapes[0] == nil {
		log.Printf("[WARN] Storage Gateway Tape With Barcode %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	tape := output.Tapes[0]

	arn := tape.TapeARN
	d.Set("arn", arn)

	d.Set("vtl_device", tape.VTLDevice)
	d.Set("kms_key_arn", tape.KMSKey)
	d.Set("pool_id", tape.PoolId)
	d.Set("tape_barcode", tape.TapeBarcode)
	d.Set("tape_size_in_bytes", int(aws.Int64Value(tape.TapeSizeInBytes)))
	d.Set("tape_used_in_bytes", int(aws.Int64Value(tape.TapeUsedInBytes)))

	tags, err := keyvaluetags.StoragegatewayListTags(conn, *arn)
	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", *arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsStorageGatewayTapeWithBarcodeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsStorageGatewayTapeWithBarcodeRead(d, meta)
}

func resourceAwsStorageGatewayTapeWithBarcodeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DeleteTapeInput{
		GatewayARN: aws.String(d.Get("gateway_arn").(string)),
		TapeARN:    aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway Tape With Barcode: %s", input)
	_, err := conn.DeleteTape(input)
	if err != nil {
		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified tape was not found.") {
			return nil
		}
		return fmt.Errorf("error deleting Storage Gateway Tape With Barcode: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:        []string{"AVAILABLE", "DELETING", "FORCE_DELETING"},
		Target:         []string{"MISSING"},
		Refresh:        storageGatewayTapeWithBarcodeRefreshFunc(d.Id(), conn),
		Timeout:        d.Timeout(schema.TimeoutDelete),
		Delay:          5 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 1,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error waiting for Storage Gateway Tape With Barcode deletion: %s", err)
	}

	return nil
}

func storageGatewayTapeWithBarcodeRefreshFunc(tapeARN string, conn *storagegateway.StorageGateway) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeTapesInput{
			TapeARNs: []*string{aws.String(tapeARN)},
		}

		log.Printf("[DEBUG] Reading Storage Gateway Tape With Barcode: %s", input)
		output, err := conn.DescribeTapes(input)
		if err != nil {
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				return nil, "MISSING", nil
			}
			return nil, "ERROR", fmt.Errorf("error reading Storage Gateway Tape With Barcode: %s", err)
		}

		if output == nil || len(output.Tapes) == 0 || output.Tapes[0] == nil {
			return nil, "MISSING", nil
		}

		tape := output.Tapes[0]

		return tape, aws.StringValue(tape.TapeStatus), nil
	}
}
