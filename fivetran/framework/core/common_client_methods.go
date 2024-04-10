package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/certificates"
	"github.com/fivetran/go-fivetran/common"
)

func RevokeCertificates(ctx context.Context, client *fivetran.Client, id, serviceType string, hashes []string) (common.CommonResponse, error) {
	var resp common.CommonResponse = common.CommonResponse{Code: "", Message: ""}
	for _, h := range hashes {
		var err error = nil
		if serviceType == "connector" {
			svc := client.NewConnectorCertificateRevoke()
			resp, err = svc.ConnectorID(id).Hash(h).Do(ctx)
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

func ReadCertificatesFromUpstream(ctx context.Context, client *fivetran.Client, id string, serviceType string) (certificates.CertificatesListResponse, error) {
	var respNextCursor string
	var listResponse certificates.CertificatesListResponse
	limit := 1000

	for {
		var err error
		var tmpResp certificates.CertificatesListResponse

		if serviceType == "connector" {
			svc := client.NewConnectorCertificatesList().ConnectorID(id).Limit(limit)
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
