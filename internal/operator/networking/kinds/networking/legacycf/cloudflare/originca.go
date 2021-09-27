package cloudflare

import (
	"context"
	"crypto/rsa"

	"github.com/caos/orbos/v5/internal/operator/networking/kinds/networking/legacycf/cloudflare/certificate"
	"github.com/cloudflare/cloudflare-go"
)

func (c *Cloudflare) CreateOriginCACertificate(ctx context.Context, domain string, hosts []string, key *rsa.PrivateKey) (*cloudflare.OriginCACertificate, error) {

	csr, err := certificate.GetCSR(domain, key)
	if err != nil {
		return nil, err
	}

	csrPem, err := certificate.PEMEncodeCSR(csr)
	if err != nil {
		return nil, err
	}

	cert := cloudflare.OriginCACertificate{
		Hostnames:       hosts,
		RequestType:     "origin-rsa",
		RequestValidity: 5475,
		CSR:             string(csrPem),
	}

	return c.api.CreateOriginCertificate(ctx, cert)
}

func (c *Cloudflare) GetOriginCACertificates(ctx context.Context, domain string) ([]cloudflare.OriginCACertificate, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	return c.api.OriginCertificates(ctx, cloudflare.OriginCACertificateListOptions{ZoneID: id})
}

func (c *Cloudflare) RevokeOriginCACertificate(ctx context.Context, id string) error {
	_, err := c.api.RevokeOriginCertificate(ctx, id)
	return err
}
