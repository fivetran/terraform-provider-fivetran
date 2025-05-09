package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/certificates"
	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/fingerprints"
)

func RevokeCertificates(ctx context.Context, client *fivetran.Client, id, serviceType string, hashes []string) (common.CommonResponse, error) {
	var resp common.CommonResponse = common.CommonResponse{Code: "", Message: ""}
	for _, h := range hashes {
		var err error = nil
		if serviceType == "connection" {
			svc := client.NewConnectionCertificateRevoke()
			resp, err = svc.ConnectionID(id).Hash(h).Do(ctx)
		}
		if serviceType == "connector" {
			svc := client.NewConnectionCertificateRevoke()
			resp, err = svc.ConnectionID(id).Hash(h).Do(ctx)
		}
		if serviceType == "destination" {
			svc := client.NewDestinationCertificateRevoke()
			resp, err = svc.DestinationID(id).Hash(h).Do(ctx)
		}
		if err != nil && !strings.HasPrefix(resp.Code, "NotFound") {
			return resp, fmt.Errorf("Unable to revoke certificate with hash = %v", h)
		}
	}
	return resp, nil
}

func RevokeFingerptints(ctx context.Context, client *fivetran.Client, id, serviceType string, hashes []string) (common.CommonResponse, error) {
	var resp common.CommonResponse = common.CommonResponse{Code: "", Message: ""}
	for _, h := range hashes {
		var err error = nil
		if serviceType == "connection" {
			svc := client.NewConnectionFingerprintRevoke()
			resp, err = svc.ConnectionID(id).Hash(h).Do(ctx)
		}
		if serviceType == "connector" {
			svc := client.NewConnectionFingerprintRevoke()
			resp, err = svc.ConnectionID(id).Hash(h).Do(ctx)
		}
		if serviceType == "destination" {
			svc := client.NewDestinationFingerprintRevoke()
			resp, err = svc.DestinationID(id).Hash(h).Do(ctx)
		}
		if err != nil && !strings.HasPrefix(resp.Code, "NotFound") {
			return resp, fmt.Errorf("Unable to revoke fingerprint with hash = %v", h)
		}
	}
	return resp, nil
}

func ReadCertificatesFromUpstream(ctx context.Context, client *fivetran.Client, id string, serviceType string) (certificates.CertificatesListResponse, error) {
	var respNextCursor string
	var listResponse certificates.CertificatesListResponse
	limit := 1000

	for {
		var err error
		var tmpResp certificates.CertificatesListResponse

		if serviceType == "connection" {
			svc := client.NewConnectionCertificatesList().ConnectionID(id).Limit(limit)
			if respNextCursor != "" {
				svc.Cursor(respNextCursor)
			}
			tmpResp, err = svc.Do(ctx)
		}

		if serviceType == "connector" {
			svc := client.NewConnectionCertificatesList().ConnectionID(id).Limit(limit)
			if respNextCursor != "" {
				svc.Cursor(respNextCursor)
			}
			tmpResp, err = svc.Do(ctx)
		}

		if serviceType == "destination" {
			svc := client.NewDestinationCertificatesList().DestinationID(id).Limit(limit)
			if respNextCursor != "" {
				svc.Cursor(respNextCursor)
			}
			tmpResp, err = svc.Do(ctx)
		}

		if err != nil {
			return certificates.CertificatesListResponse{}, err
		}

		listResponse.Data.Items = append(listResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	return listResponse, nil
}

func ReadFromSourceFingerprintCommon(ctx context.Context, client *fivetran.Client, id string, serviceType string) (fingerprints.FingerprintsListResponse, error) {
	var respNextCursor string
	var listResponse fingerprints.FingerprintsListResponse
	var err error
	limit := 1000

	for {
		var tmpResp fingerprints.FingerprintsListResponse

		if serviceType == "connection" {
			svc := client.NewConnectionFingerprintsList().ConnectionID(id).Limit(limit)
			if respNextCursor != "" {
				svc.Cursor(respNextCursor)
			}
			tmpResp, err = svc.Do(ctx)
		}

		if serviceType == "connector" {
			svc := client.NewConnectionFingerprintsList().ConnectionID(id).Limit(limit)
			if respNextCursor != "" {
				svc.Cursor(respNextCursor)
			}
			tmpResp, err = svc.Do(ctx)
		}

		if serviceType == "destination" {
			svc := client.NewDestinationFingerprintsList().DestinationID(id).Limit(limit)
			if respNextCursor != "" {
				svc.Cursor(respNextCursor)
			}
			tmpResp, err = svc.Do(ctx)
		}

		if err != nil {
			listResponse = fingerprints.FingerprintsListResponse{}
			return listResponse, err
		}

		listResponse.Data.Items = append(listResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	return listResponse, nil
}
