package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/certificates"
	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCertificates(resourceType ResourceType) *schema.Resource {
	return &schema.Resource{
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceCertificatesRead(ctx, d, m, resourceType, "certificates")
			if err != nil {
				return helpers.NewDiagAppend(diags, diag.Error, "read error", err.Error())
			}
			return diags
		},
		Schema: resourceCertificatesSchema(true, resourceType),
	}
}

func resourceCertificates(resourceType ResourceType) *schema.Resource {
	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceCertificatesCreate(ctx, d, m, resourceType)
			if err != nil {
				return helpers.NewDiagAppend(diags, diag.Error, "create error", err.Error())
			}
			return diags
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceCertificatesRead(ctx, d, m, resourceType, "certificate")
			if err != nil {
				return helpers.NewDiagAppend(diags, diag.Error, "read error", err.Error())
			}
			return diags
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceCertificatesUpdate(ctx, d, m, resourceType)
			if err != nil {
				return helpers.NewDiagAppend(diags, diag.Error, "update error", err.Error())
			}
			return diags
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceCertificatesDelete(ctx, d, m, resourceType)
			if err != nil {
				return helpers.NewDiagAppend(diags, diag.Error, "delete error", err.Error())
			}
			return diags
		},
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:   resourceCertificatesSchema(false, resourceType),
	}
}

func resourceCertificatesSchema(datasource bool, resourceType ResourceType) map[string]*schema.Schema {
	itemsField := "certificate"
	if datasource {
		itemsField = itemsField + "s"
	}
	result := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    !datasource,
			Required:    datasource,
			Description: "The unique identifier for the resource. Equal to target " + resourceType.String() + " id.",
		},
		resourceType.String() + "_id": {
			Type:        schema.TypeString,
			Computed:    datasource,
			Required:    !datasource,
			ForceNew:    !datasource,
			Description: "The unique identifier for the target " + resourceType.String() + " within the Fivetran system.",
		},
		itemsField: certificateSchema(datasource),
	}
	return result
}

func certificateSchema(datasource bool) *schema.Schema {
	elemSchema := map[string]*schema.Schema{
		"hash": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "Hash of the fingerprint.",
		},

		"public_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Certificate public key.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Certificate name.",
		},
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Certificate type.",
		},
		"sha1": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Certificate sha1.",
		},
		"sha256": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Certificate sha256.",
		},
		"validated_by": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "User name who validated the certificate.",
		},
		"validated_date": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date when the certificate was approved.",
		},
	}

	if !datasource {
		elemSchema["encoded_cert"] = &schema.Schema{
			Type:        schema.TypeString,
			Sensitive:   true,
			Required:    true,
			Description: "Base64 encoded certificate.",
		}
	}
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: fingerprintItemHash,
		Elem: &schema.Resource{
			Schema: elemSchema,
		},
	}
}

func resourceCertificatesCreate(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType) error {
	client := m.(*fivetran.Client)
	// Verify resource exists
	var id = d.Get(resourceType.String() + "_id").(string)
	resourceId, err := verifyApproveTarget(ctx, client, id, resourceType)
	if err != nil {
		return err
	}
	// Sync certificates
	err = syncCertificates(ctx, client, d.Get("certificate").(*schema.Set).List(), id, resourceType)
	if err != nil {
		return err
	}
	d.SetId(resourceId)
	return resourceCertificatesRead(ctx, d, m, resourceType, "certificate")
}

func resourceCertificatesRead(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType, itemsField string) error {
	id := d.Get("id").(string)
	response, err := fetchCertificates(ctx, m.(*fivetran.Client), id, resourceType)
	if err != nil {
		return fmt.Errorf("%v; code: %v; message: %v", err, response.Code, response.Message)
	}

	local := mapItemsFromResourceData(d.Get(itemsField).(*schema.Set).List())
	msi := make(map[string]interface{})
	msi[resourceType.String()+"_id"] = id
	msi[itemsField] = flattenCertificates(response, local)
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	d.SetId(id)
	return nil
}

func resourceCertificatesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType) error {
	client := m.(*fivetran.Client)
	id := d.Get(resourceType.String() + "_id").(string)
	if d.HasChange("certificate") {
		// Sync fingerprints
		err := syncCertificates(ctx, client, d.Get("certificate").(*schema.Set).List(), id,
			resourceType,
		)
		if err != nil {
			return err
		}
	}
	return resourceCertificatesRead(ctx, d, m, resourceType, "certificate")
}

func resourceCertificatesDelete(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType) error {
	client := m.(*fivetran.Client)
	id := d.Get(resourceType.String() + "_id").(string)

	// Sync with empty local fingerprints list leads to cleanup
	return syncCertificates(ctx, client, []interface{}{}, id,
		resourceType)
}

func syncCertificates(ctx context.Context, client *fivetran.Client, local []interface{}, id string, resourceType ResourceType) error {
	response, err := fetchCertificates(ctx, client, id, resourceType)
	if err == nil {
		upstream := make([]string, 0)
		for k := range mapCertificatesFromResponse(response) {
			upstream = append(upstream, k)
		}
		localItems := mapItemsFromResourceData(local)
		local := make([]string, 0)
		for k := range localItems {
			local = append(local, k)
		}
		revoke, _, approve := helpers.Intersection(upstream, local)

		for _, r := range revoke {
			resp, err := revokeCertificate(ctx, client, id, r, resourceType)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
		for _, a := range approve {
			resp, err := approveCertificate(ctx, client, id, a, localItems[a].(map[string]interface{})["encoded_cert"].(string), resourceType)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
	}
	return err
}

func fetchCertificates(ctx context.Context, client *fivetran.Client, id string, resourceType ResourceType) (certificates.CertificatesListResponse, error) {
	var result certificates.CertificatesListResponse
	result, err := getCertificates(ctx, client, id, "", resourceType)
	cursor := result.Data.NextCursor
	for err == nil && cursor != "" {
		if innerResult, err := getCertificates(ctx, client, id, cursor, resourceType); err == nil {
			cursor = innerResult.Data.NextCursor
			result.Data.Items = append(result.Data.Items, innerResult.Data.Items...)
		}
	}
	return result, err
}

func mapCertificatesFromResponse(resp certificates.CertificatesListResponse) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fp := range resp.Data.Items {
		result[fp.Hash] = flattenCertificate(fp)
	}
	return result
}

func flattenCertificates(resp certificates.CertificatesListResponse, local map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)
	for _, fc := range resp.Data.Items {
		cert := flattenCertificate(fc)
		// copy local value from resource data to state
		if localCertificate, ok := local[fc.Hash].(map[string]interface{}); ok {
			cert["encoded_cert"] = localCertificate["encoded_cert"]
		}
		result = append(result, cert)
	}
	return result
}

func flattenCertificate(resp certificates.CertificateDetails) map[string]interface{} {
	f := make(map[string]interface{})
	f["hash"] = resp.Hash
	f["public_key"] = resp.PublicKey
	f["name"] = resp.Name
	f["type"] = resp.Type
	f["sha1"] = resp.Sha1
	f["sha256"] = resp.Sha256
	f["validated_by"] = resp.ValidatedBy
	f["validated_date"] = resp.ValidatedDate
	return f
}

func getCertificates(ctx context.Context, client *fivetran.Client, id, cursor string, resourceType ResourceType) (certificates.CertificatesListResponse, error) {
	switch resourceType {
	case Connector:
		return getConnectorCertificatesFunction(ctx, client, id, cursor)
	case Destination:
		return getDestinationCertificatesFunction(ctx, client, id, cursor)
	}
	var result certificates.CertificatesListResponse
	return result, fmt.Errorf("unknown resource type %v", resourceType.String())
}

func approveCertificate(ctx context.Context, client *fivetran.Client, id, hash, encodedCert string, resourceType ResourceType) (certificates.CertificateResponse, error) {
	switch resourceType {
	case Connector:
		return approveConnectorCertificateFunc(ctx, client, id, hash, encodedCert)
	case Destination:
		return approveDestinationCertificateFunc(ctx, client, id, hash, encodedCert)
	}
	var result certificates.CertificateResponse
	return result, fmt.Errorf("unknown resource type %v", resourceType.String())
}

func revokeCertificate(ctx context.Context, client *fivetran.Client, id, hash string, resourceType ResourceType) (common.CommonResponse, error) {
	switch resourceType {
	case Connector:
		return revokeConnectorCertificateFunc(ctx, client, id, hash)
	case Destination:
		return revokeDestinationCertificateFunc(ctx, client, id, hash)
	}
	var result common.CommonResponse
	return result, fmt.Errorf("unknown resource type %v", resourceType.String())
}

// Connectors
func getConnectorCertificatesFunction(ctx context.Context, client *fivetran.Client, id, cursor string) (certificates.CertificatesListResponse, error) {
	svc := client.NewConnectorCertificatesList().ConnectorID(id)
	if cursor != "" {
		svc.Cursor(cursor)
	}
	return svc.Do(ctx)
}

func approveConnectorCertificateFunc(ctx context.Context, client *fivetran.Client, id, hash, encodedCert string) (certificates.CertificateResponse, error) {
	return client.NewCertificateConnectorCertificateApprove().
		ConnectorID(id).
		Hash(hash).
		EncodedCert(encodedCert).
		Do(ctx)
}

func revokeConnectorCertificateFunc(ctx context.Context, client *fivetran.Client, id, hash string) (common.CommonResponse, error) {
	return client.NewConnectorCertificateRevoke().
		ConnectorID(id).
		Hash(hash).
		Do(ctx)
}

// Destinations
func getDestinationCertificatesFunction(ctx context.Context, client *fivetran.Client, id, cursor string) (certificates.CertificatesListResponse, error) {
	svc := client.NewDestinationCertificatesList().DestinationID(id)
	if cursor != "" {
		svc.Cursor(cursor)
	}
	return svc.Do(ctx)
}

func approveDestinationCertificateFunc(ctx context.Context, client *fivetran.Client, id, hash, encodedCert string) (certificates.CertificateResponse, error) {
	return client.NewCertificateDestinationCertificateApprove().
		DestinationID(id).
		Hash(hash).
		EncodedCert(encodedCert).
		Do(ctx)
}

func revokeDestinationCertificateFunc(ctx context.Context, client *fivetran.Client, id, hash string) (common.CommonResponse, error) {
	return client.NewDestinationCertificateRevoke().
		DestinationID(id).
		Hash(hash).
		Do(ctx)
}
