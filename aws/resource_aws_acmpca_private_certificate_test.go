package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAcmpcaPrivateCertificate_Basic(t *testing.T) {
	resourceName := "aws_acmpca_private_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsAcmpcaPrivateCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaPrivateCertificateConfig_Required,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaPrivateCertificateExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:acm-pca:[^:]+:[^:]+:certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "validity_length", "1"),
					resource.TestCheckResourceAttr(resourceName, "validity_unit", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					resource.TestCheckResourceAttr(resourceName, "template_arn", "arn:aws:acm-pca:::template/EndEntityCertificate/V1"),
				),
			},
		},
	})
}

func testAccCheckAwsAcmpcaPrivateCertificateDestroy(s *terraform.State) error {
	// unfortunately aws pca does not have an API to determine if a cert has been revoked.
	// see: https://docs.aws.amazon.com/acm-pca/latest/userguide/PcaRevokeCert.html
	return nil
}

func testAccCheckAwsAcmpcaPrivateCertificateExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).acmpcaconn
		input := &acmpca.GetCertificateInput{
			CertificateArn:          aws.String(rs.Primary.ID),
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.GetCertificate(input)

		if err != nil {
			return err
		}

		if output == nil || output.Certificate == nil {
			return fmt.Errorf("ACMPCA Certificate %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

const testAccAwsAcmpcaPrivateCertificateConfig_Required = `
resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_cert_request" "csr" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "testing"
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_private_certificate" "test" {
	certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
	certificate_signing_request = tls_cert_request.csr.cert_request_pem
	signing_algorithm = "SHA256WITHRSA"
	validity_length = 1
	validity_unit = "YEARS"
}
`
