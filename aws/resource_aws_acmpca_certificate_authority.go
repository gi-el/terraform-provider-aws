package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAcmpcaCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmpcaCertificateAuthorityCreate,
		Read:   resourceAwsAcmpcaCertificateAuthorityRead,
		Update: resourceAwsAcmpcaCertificateAuthorityUpdate,
		Delete: resourceAwsAcmpcaCertificateAuthorityDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("permanent_deletion_time_in_days", 30)

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
		MigrateState:  resourceAwsAcmpcaCertificateAuthorityMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_CertificateAuthorityConfiguration.html
			"certificate_authority_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_algorithm": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								acmpca.KeyAlgorithmEcPrime256v1,
								acmpca.KeyAlgorithmEcSecp384r1,
								acmpca.KeyAlgorithmRsa2048,
								acmpca.KeyAlgorithmRsa4096,
							}, false),
						},
						"signing_algorithm": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								acmpca.SigningAlgorithmSha256withecdsa,
								acmpca.SigningAlgorithmSha256withrsa,
								acmpca.SigningAlgorithmSha384withecdsa,
								acmpca.SigningAlgorithmSha384withrsa,
								acmpca.SigningAlgorithmSha512withecdsa,
								acmpca.SigningAlgorithmSha512withrsa,
							}, false),
						},
						// https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_ASN1Subject.html
						"subject": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"common_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"country": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 2),
									},
									"distinguished_name_qualifier": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"generation_qualifier": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 3),
									},
									"given_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 16),
									},
									"initials": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 5),
									},
									"locality": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"organization": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"organizational_unit": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"pseudonym": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"state": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"surname": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 40),
									},
									"title": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
								},
							},
						},
					},
				},
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_RevocationConfiguration.html
			"revocation_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_CrlConfiguration.html
						"crl_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == "1" && new == "0" {
									return true
								}
								return false
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"custom_cname": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 253),
									},
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									// ValidationException: 1 validation error detected: Value null or empty at 'expirationInDays' failed to satisfy constraint: Member must not be null or empty.
									// InvalidParameter: 1 validation error(s) found. minimum field value of 1, CreateCertificateAuthorityInput.RevocationConfiguration.CrlConfiguration.ExpirationInDays.
									"expiration_in_days": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 5000),
									},
									"s3_bucket_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
								},
							},
						},
					},
				},
			},
			"serial": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permanent_deletion_time_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(7, 30),
			},
			"tags": tagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  acmpca.CertificateAuthorityTypeSubordinate,
				ValidateFunc: validation.StringInSlice([]string{
					acmpca.CertificateAuthorityTypeRoot,
					acmpca.CertificateAuthorityTypeSubordinate,
				}, false),
			},
			"validity_length": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"validity_unit": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					acmpca.ValidityPeriodTypeAbsolute,
					acmpca.ValidityPeriodTypeDays,
					acmpca.ValidityPeriodTypeEndDate,
					acmpca.ValidityPeriodTypeMonths,
					acmpca.ValidityPeriodTypeYears,
				}, false),
			},
		},
	}
}

func resourceAwsAcmpcaCertificateAuthorityCreate(d *schema.ResourceData, meta interface{}) error {
	isRootCA := d.Get("type").(string) == acmpca.CertificateAuthorityTypeRoot

	if isRootCA {
		if _, ok := d.GetOk("validity_length"); !ok {
			return fmt.Errorf("validity_length must be set when creating a Certificate Authority with a self-signed root certificate")
		}
		if _, ok := d.GetOk("validity_unit"); !ok {
			return fmt.Errorf("validity_unit must be set when creating a Certificate Authority with a self-signed root certificate")
		}
	}

	conn := meta.(*AWSClient).acmpcaconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AcmpcaTags()

	input := &acmpca.CreateCertificateAuthorityInput{
		CertificateAuthorityConfiguration: expandAcmpcaCertificateAuthorityConfiguration(d.Get("certificate_authority_configuration").([]interface{})),
		CertificateAuthorityType:          aws.String(d.Get("type").(string)),
		IdempotencyToken:                  aws.String(resource.UniqueId()),
		RevocationConfiguration:           expandAcmpcaRevocationConfiguration(d.Get("revocation_configuration").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = tags
	}

	log.Printf("[DEBUG] Creating ACMPCA Certificate Authority: %s", input)
	var output *acmpca.CreateCertificateAuthorityOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.CreateCertificateAuthority(input)
		if err != nil {
			// ValidationException: The ACM Private CA service account 'acm-pca-prod-pdx' requires getBucketAcl permissions for your S3 bucket 'tf-acc-test-5224996536060125340'. Check your S3 bucket permissions and try again.
			if isAWSErr(err, "ValidationException", "Check your S3 bucket permissions and try again") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.CreateCertificateAuthority(input)
	}
	if err != nil {
		return fmt.Errorf("error creating ACMPCA Certificate Authority: %s", err)
	}

	d.SetId(aws.StringValue(output.CertificateAuthorityArn))

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"",
			acmpca.CertificateAuthorityStatusCreating,
		},
		Target: []string{
			acmpca.CertificateAuthorityStatusActive,
			acmpca.CertificateAuthorityStatusPendingCertificate,
		},
		Refresh: acmpcaCertificateAuthorityRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for ACMPCA Certificate Authority %q to be active or pending certificate: %s", d.Id(), err)
	}

	// If we request a ROOT CA, we'll ask AWS to self-sign right away
	if isRootCA {
		err = acmpcaCertificateAuthoritySelfSignCert(conn, d)
		if err != nil {
			return fmt.Errorf("error self-signing ACMPCA CA Certificate for CA %q of type ROOT: %s", d.Id(), err)
		}
	}

	return resourceAwsAcmpcaCertificateAuthorityRead(d, meta)
}

func resourceAwsAcmpcaCertificateAuthorityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	describeCertificateAuthorityInput := &acmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority: %s", describeCertificateAuthorityInput)

	describeCertificateAuthorityOutput, err := conn.DescribeCertificateAuthority(describeCertificateAuthorityInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACMPCA Certificate Authority %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading ACMPCA Certificate Authority: %s", err)
	}

	if describeCertificateAuthorityOutput.CertificateAuthority == nil {
		log.Printf("[WARN] ACMPCA Certificate Authority %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}
	certificateAuthority := describeCertificateAuthorityOutput.CertificateAuthority

	d.Set("arn", certificateAuthority.Arn)

	if err := d.Set("certificate_authority_configuration", flattenAcmpcaCertificateAuthorityConfiguration(certificateAuthority.CertificateAuthorityConfiguration)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("enabled", (aws.StringValue(certificateAuthority.Status) != acmpca.CertificateAuthorityStatusDisabled))
	d.Set("not_after", aws.TimeValue(certificateAuthority.NotAfter).Format(time.RFC3339))
	d.Set("not_before", aws.TimeValue(certificateAuthority.NotBefore).Format(time.RFC3339))

	if err := d.Set("revocation_configuration", flattenAcmpcaRevocationConfiguration(certificateAuthority.RevocationConfiguration)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("serial", certificateAuthority.Serial)
	d.Set("status", certificateAuthority.Status)
	d.Set("type", certificateAuthority.Type)

	getCertificateAuthorityCertificateInput := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority Certificate: %s", getCertificateAuthorityCertificateInput)

	getCertificateAuthorityCertificateOutput, err := conn.GetCertificateAuthorityCertificate(getCertificateAuthorityCertificateInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACMPCA Certificate Authority %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		// Returned when in PENDING_CERTIFICATE status
		// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
		if !isAWSErr(err, acmpca.ErrCodeInvalidStateException, "") {
			return fmt.Errorf("error reading ACMPCA Certificate Authority Certificate: %s", err)
		}
	}

	d.Set("certificate", "")
	d.Set("certificate_chain", "")
	if getCertificateAuthorityCertificateOutput != nil {
		d.Set("certificate", getCertificateAuthorityCertificateOutput.Certificate)
		d.Set("certificate_chain", getCertificateAuthorityCertificateOutput.CertificateChain)
	}

	getCertificateAuthorityCsrInput := &acmpca.GetCertificateAuthorityCsrInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority Certificate Signing Request: %s", getCertificateAuthorityCsrInput)

	getCertificateAuthorityCsrOutput, err := conn.GetCertificateAuthorityCsr(getCertificateAuthorityCsrInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACMPCA Certificate Authority %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if !isAWSErr(err, acmpca.ErrCodeInvalidStateException, "") {
			return fmt.Errorf("error reading ACMPCA Certificate Authority Certificate Signing Request: %s", err)
		}
	}

	d.Set("certificate_signing_request", "")
	if getCertificateAuthorityCsrOutput != nil {
		d.Set("certificate_signing_request", getCertificateAuthorityCsrOutput.Csr)
	}

	tags, err := keyvaluetags.AcmpcaListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for ACMPCA Certificate Authority (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAcmpcaCertificateAuthorityUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn
	updateCertificateAuthority := false

	input := &acmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	if d.HasChange("enabled") {
		input.Status = aws.String(acmpca.CertificateAuthorityStatusActive)
		if !d.Get("enabled").(bool) {
			input.Status = aws.String(acmpca.CertificateAuthorityStatusDisabled)
		}
		updateCertificateAuthority = true
	}

	if d.HasChange("revocation_configuration") {
		input.RevocationConfiguration = expandAcmpcaRevocationConfiguration(d.Get("revocation_configuration").([]interface{}))
		updateCertificateAuthority = true
	}

	if updateCertificateAuthority {
		log.Printf("[DEBUG] Updating ACMPCA Certificate Authority: %s", input)
		_, err := conn.UpdateCertificateAuthority(input)
		if err != nil {
			return fmt.Errorf("error updating ACMPCA Certificate Authority: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AcmpcaUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating ACMPCA Certificate Authority (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsAcmpcaCertificateAuthorityRead(d, meta)
}

func resourceAwsAcmpcaCertificateAuthorityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	input := &acmpca.DeleteCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	if v, exists := d.GetOk("permanent_deletion_time_in_days"); exists {
		input.PermanentDeletionTimeInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Deleting ACMPCA Certificate Authority: %s", input)
	_, err := conn.DeleteCertificateAuthority(input)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting ACMPCA Certificate Authority: %s", err)
	}

	return nil
}

func acmpcaCertificateAuthorityRefreshFunc(conn *acmpca.ACMPCA, certificateAuthorityArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(certificateAuthorityArn),
		}

		log.Printf("[DEBUG] Reading ACMPCA Certificate Authority: %s", input)
		output, err := conn.DescribeCertificateAuthority(input)
		if err != nil {
			if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
				return nil, "", nil
			}
			return nil, "", err
		}

		if output == nil || output.CertificateAuthority == nil {
			return nil, "", nil
		}

		return output.CertificateAuthority, aws.StringValue(output.CertificateAuthority.Status), nil
	}
}

func acmpcaCertificateAuthoritySelfSignCert(conn *acmpca.ACMPCA, d *schema.ResourceData) error {
	// Get CSR
	getCertificateAuthorityCsrInput := &acmpca.GetCertificateAuthorityCsrInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority Certificate Signing Request: %s", getCertificateAuthorityCsrInput)

	getCertificateAuthorityCsrOutput, err := conn.GetCertificateAuthorityCsr(getCertificateAuthorityCsrInput)
	if err != nil {
		return fmt.Errorf("error reading ACMPCA Certificate Authority Certificate Signing Request: %s", err)
	}

	csr := aws.StringValue(getCertificateAuthorityCsrOutput.Csr)

	certificateAuthorityConfiguration := d.Get("certificate_authority_configuration").(map[string]interface{})

	// Issue self-signed certificate
	issueCertificateInput := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String(d.Id()),
		Csr:                     []byte(csr),
		IdempotencyToken:        aws.String(resource.UniqueId()),
		SigningAlgorithm:        aws.String(certificateAuthorityConfiguration["signing_algorithm"].(string)),
		TemplateArn:             aws.String("arn:aws:acm-pca:::template/RootCACertificate/V1"),
		Validity: &acmpca.Validity{
			Type:  aws.String(d.Get("validity_unit").(string)),
			Value: aws.Int64(int64(d.Get("validity_length").(int))),
		},
	}

	log.Printf("[DEBUG] ACMPCA Issue Certificate: %s", issueCertificateInput)

	issueCertificateOutput, err := conn.IssueCertificate(issueCertificateInput)
	if err != nil {
		return fmt.Errorf("error issuing ACMPCA Certificate: %s", err)
	}

	certificateArn := aws.StringValue(issueCertificateOutput.CertificateArn)

	// Wait until certificate is ready
	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(certificateArn),
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] ACMPCA Get Certificate: %s", getCertificateInput)

	err = conn.WaitUntilCertificateIssued(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error waiting for ACMPCA to issue Certificate %q: %s", certificateArn, err)
	}

	getCertificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error retrieving ACMPCA Certificate %q: %s", certificateArn, err)
	}

	// Import certificate to CA
	importCertificateAuthorityCertificateInput := &acmpca.ImportCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(d.Id()),
		Certificate:             []byte(aws.StringValue(getCertificateOutput.Certificate)),
	}

	log.Printf("[DEBUG] ACMPCA import Certificate Authority Certificate: %s", importCertificateAuthorityCertificateInput)

	_, err = conn.ImportCertificateAuthorityCertificate(importCertificateAuthorityCertificateInput)
	if err != nil {
		return fmt.Errorf("error importing ACMPCA Certificate Authority Certificate %q in ACMPCA Certificate Authority %q: %s", certificateArn, d.Id(), err)
	}

	return nil
}

func expandAcmpcaASN1Subject(l []interface{}) *acmpca.ASN1Subject {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	subject := &acmpca.ASN1Subject{}
	if v, ok := m["common_name"]; ok && v.(string) != "" {
		subject.CommonName = aws.String(v.(string))
	}
	if v, ok := m["country"]; ok && v.(string) != "" {
		subject.Country = aws.String(v.(string))
	}
	if v, ok := m["distinguished_name_qualifier"]; ok && v.(string) != "" {
		subject.DistinguishedNameQualifier = aws.String(v.(string))
	}
	if v, ok := m["generation_qualifier"]; ok && v.(string) != "" {
		subject.GenerationQualifier = aws.String(v.(string))
	}
	if v, ok := m["given_name"]; ok && v.(string) != "" {
		subject.GivenName = aws.String(v.(string))
	}
	if v, ok := m["initials"]; ok && v.(string) != "" {
		subject.Initials = aws.String(v.(string))
	}
	if v, ok := m["locality"]; ok && v.(string) != "" {
		subject.Locality = aws.String(v.(string))
	}
	if v, ok := m["organization"]; ok && v.(string) != "" {
		subject.Organization = aws.String(v.(string))
	}
	if v, ok := m["organizational_unit"]; ok && v.(string) != "" {
		subject.OrganizationalUnit = aws.String(v.(string))
	}
	if v, ok := m["pseudonym"]; ok && v.(string) != "" {
		subject.Pseudonym = aws.String(v.(string))
	}
	if v, ok := m["state"]; ok && v.(string) != "" {
		subject.State = aws.String(v.(string))
	}
	if v, ok := m["surname"]; ok && v.(string) != "" {
		subject.Surname = aws.String(v.(string))
	}
	if v, ok := m["title"]; ok && v.(string) != "" {
		subject.Title = aws.String(v.(string))
	}

	return subject
}

func expandAcmpcaCertificateAuthorityConfiguration(l []interface{}) *acmpca.CertificateAuthorityConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.CertificateAuthorityConfiguration{
		KeyAlgorithm:     aws.String(m["key_algorithm"].(string)),
		SigningAlgorithm: aws.String(m["signing_algorithm"].(string)),
		Subject:          expandAcmpcaASN1Subject(m["subject"].([]interface{})),
	}

	return config
}

func expandAcmpcaCrlConfiguration(l []interface{}) *acmpca.CrlConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.CrlConfiguration{
		Enabled: aws.Bool(m["enabled"].(bool)),
	}

	if v, ok := m["custom_cname"]; ok && v.(string) != "" {
		config.CustomCname = aws.String(v.(string))
	}
	if v, ok := m["expiration_in_days"]; ok && v.(int) > 0 {
		config.ExpirationInDays = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["s3_bucket_name"]; ok && v.(string) != "" {
		config.S3BucketName = aws.String(v.(string))
	}

	return config
}

func expandAcmpcaRevocationConfiguration(l []interface{}) *acmpca.RevocationConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.RevocationConfiguration{
		CrlConfiguration: expandAcmpcaCrlConfiguration(m["crl_configuration"].([]interface{})),
	}

	return config
}

func flattenAcmpcaASN1Subject(subject *acmpca.ASN1Subject) []interface{} {
	if subject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"common_name":                  aws.StringValue(subject.CommonName),
		"country":                      aws.StringValue(subject.Country),
		"distinguished_name_qualifier": aws.StringValue(subject.DistinguishedNameQualifier),
		"generation_qualifier":         aws.StringValue(subject.GenerationQualifier),
		"given_name":                   aws.StringValue(subject.GivenName),
		"initials":                     aws.StringValue(subject.Initials),
		"locality":                     aws.StringValue(subject.Locality),
		"organization":                 aws.StringValue(subject.Organization),
		"organizational_unit":          aws.StringValue(subject.OrganizationalUnit),
		"pseudonym":                    aws.StringValue(subject.Pseudonym),
		"state":                        aws.StringValue(subject.State),
		"surname":                      aws.StringValue(subject.Surname),
		"title":                        aws.StringValue(subject.Title),
	}

	return []interface{}{m}
}

func flattenAcmpcaCertificateAuthorityConfiguration(config *acmpca.CertificateAuthorityConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"key_algorithm":     aws.StringValue(config.KeyAlgorithm),
		"signing_algorithm": aws.StringValue(config.SigningAlgorithm),
		"subject":           flattenAcmpcaASN1Subject(config.Subject),
	}

	return []interface{}{m}
}

func flattenAcmpcaCrlConfiguration(config *acmpca.CrlConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"custom_cname":       aws.StringValue(config.CustomCname),
		"enabled":            aws.BoolValue(config.Enabled),
		"expiration_in_days": int(aws.Int64Value(config.ExpirationInDays)),
		"s3_bucket_name":     aws.StringValue(config.S3BucketName),
	}

	return []interface{}{m}
}

func flattenAcmpcaRevocationConfiguration(config *acmpca.RevocationConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"crl_configuration": flattenAcmpcaCrlConfiguration(config.CrlConfiguration),
	}

	return []interface{}{m}
}
