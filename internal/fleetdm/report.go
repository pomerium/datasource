package fleetdm

import (
	"archive/zip"
	"context"
	"fmt"
	"io"

	"github.com/pomerium/datasource/internal/jsonutil"
)

const (
	typeCertificateSHA1Fingerprint = "fleetdm.com/CertificateSHA1Fingerprint"
	typeHost                       = "fleetdm.com/Host"
)

func (srv *server) writeRecords(
	ctx context.Context,
	dst io.Writer,
) error {
	zw := zip.NewWriter(dst)

	fw, err := zw.Create(typeCertificateSHA1Fingerprint + ".json")
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	certs, err := srv.client.QueryCertificates(ctx, srv.cfg.certificateQueryID)
	if err != nil {
		return fmt.Errorf("query certificates: %w", err)
	}

	err = jsonutil.StreamWriteArray(fw, certs)
	if err != nil {
		return fmt.Errorf("write certificates: %w", err)
	}

	fw, err = zw.Create(typeHost + ".json")
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	hosts, err := srv.client.ListHosts(ctx)
	if err != nil {
		return fmt.Errorf("list hosts: %w", err)
	}

	err = jsonutil.StreamWriteArray(fw, hosts)
	if err != nil {
		return fmt.Errorf("write hosts: %w", err)
	}

	return zw.Close()
}
