package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsAcmpcaCertificateAuthorityCertificateAttachment() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentCreate,
		Read:          resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentRead,
		Delete:        resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentDelete,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"certificate_authority_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_body": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: normalizeCert,
			},
			"certificate_chain": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				StateFunc: normalizeCert,
			},
		},
	}
}

func resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn
	caARN := d.Get("certificate_authority_arn").(string)

	// Attach the certificate to the CA by importing it
	input := &acmpca.ImportCertificateAuthorityCertificateInput{
		Certificate:             []byte(d.Get("certificate_body").(string)),
		CertificateAuthorityArn: aws.String(caARN),
	}

	if chain, ok := d.GetOk("certificate_chain"); ok {
		input.CertificateChain = []byte(chain.(string))
	}

	log.Printf("[DEBUG] Importing ACMPCA Certificate Authority Certificate: %s", input)
	_, err := conn.ImportCertificateAuthorityCertificate(input)
	if err != nil {
		return fmt.Errorf("Error importing ACMPCA Certificate Authority Certificate: %s", err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", caARN)))
	return resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentRead(d, meta)
}

func resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn
	caARN := d.Get("certificate_authority_arn").(string)

	getCertificateAuthorityCertificateInput := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(caARN),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority Certificate: %s", getCertificateAuthorityCertificateInput)

	getCertificateAuthorityCertificateOutput, err := conn.GetCertificateAuthorityCertificate(getCertificateAuthorityCertificateInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACMPCA Certificate Authority %q not found - removing attachment %q from state", caARN, d.Id())
			d.SetId("")
			return nil
		}
		// Returned when in PENDING_CERTIFICATE status
		if isAWSErr(err, acmpca.ErrCodeInvalidStateException, "") {
			log.Printf("[WARN] ACMPCA Certificate Authority %q s PENDING_CERTIFICATE - removing attachment %q from state", caARN, d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading ACMPCA Certificate Authority Certificate from ACMPCA Certificate Authority %q: %s", caARN, err)
	}

	d.Set("certificate", getCertificateAuthorityCertificateOutput.Certificate)
	d.Set("certificate_chain", getCertificateAuthorityCertificateOutput.CertificateChain)

	return nil
}

func resourceAwsAcmpcaCertificateAuthorityCertificateAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	caARN := d.Get("certificate_authority_arn").(string)
	log.Printf("[WARN] Certificate Authority Certificate can never be detached from an ACMPCA Certificate Authority %q, only overwritten", caARN)
	return nil
}
