package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/fingerprints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceType int64

const (
	Connector ResourceType = iota
	Destination
)

func (lang ResourceType) String() string {
	return [...]string{
		"connector",
		"destination",
	}[lang]
}

// common function types
type getFingerprintsListFunc func(ctx context.Context, client *fivetran.Client, id, cursor string) (fingerprints.FingerprintsListResponse, error)
type approveFingerprintFunc func(ctx context.Context, client *fivetran.Client, id, hash, public_key string) (fingerprints.FingerprintResponse, error)
type revokeFingerprintFunc func(ctx context.Context, client *fivetran.Client, id, hash string) (common.CommonResponse, error)

func resourceFingerprints(resourceType ResourceType) *schema.Resource {
	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceFingerprintsCreate(ctx, d, m, resourceType)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "create error", err.Error())
			}
			return diags
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceFingerprintsRead(ctx, d, m, resourceType)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "read error", err.Error())
			}
			return diags
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceFingerprintsUpdate(ctx, d, m, resourceType)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "update error", err.Error())
			}
			return diags
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceFingerprintsDelete(ctx, d, m, resourceType)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "update error", err.Error())
			}
			return diags
		},
		Schema: resourceFingerprintsSchema(false, Connector),
	}
}

func dataSourceFingerprints(resourceType ResourceType) *schema.Resource {
	return &schema.Resource{
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			var diags diag.Diagnostics
			err := resourceFingerprintsRead(ctx, d, m, resourceType)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "read error", err.Error())
			}
			return diags
		},
		Schema: resourceFingerprintsSchema(true, Connector),
	}
}

func resourceFingerprintsSchema(datasource bool, resourceType ResourceType) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    !datasource,
			Required:    datasource,
			Description: "The unique identifier for the resource.",
		},
		resourceType.String() + "_id": {
			Type:        schema.TypeString,
			Computed:    datasource,
			Required:    !datasource,
			ForceNew:    !datasource,
			Description: "The unique identifier for the target " + resourceType.String() + " within the Fivetran system.",
		},
		"fingerprint": fingerprintSchema(datasource),
	}
	return result
}

func fingerprintSchema(datasource bool) *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: fingerprintItemHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"hash": {
					Type:        schema.TypeString,
					Required:    !datasource,
					Computed:    datasource,
					Description: "Hash of the fingerprint.",
				},
				"public_key": {
					Type:        schema.TypeString,
					Required:    !datasource,
					Computed:    datasource,
					Description: "The SSH public key.",
				},
				"validated_by": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "User name who validated the fingerprint.",
				},
				"validated_date": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The date when SSH fingerprint was approved.",
				},
			},
		},
	}
}

func resourceFingerprintsCreate(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType) error {
	switch resourceType {
	case Connector:
		err := resourceFingerprintsCreateBase(ctx, d, m, resourceType,
			getConnectorFingerprintsFunction,
			approveConnectorFingerprintFunc,
			revokeConnectorFingerprintFunc)
		if err != nil {
			return err
		}
		return resourceFingerprintsRead(ctx, d, m, resourceType)
	case Destination:
		err := resourceFingerprintsCreateBase(ctx, d, m, resourceType,
			getDestinationFingerprintsFunction,
			approveDestinationFingerprintFunc,
			revokeDestinationFingerprintFunc)
		if err != nil {
			return err
		}
		return resourceFingerprintsRead(ctx, d, m, resourceType)
	}
	return fmt.Errorf("unknown resource type " + resourceType.String())
}

func resourceFingerprintsCreateBase(ctx context.Context, d *schema.ResourceData, m interface{},
	resourceType ResourceType,
	getFingerprintsFunction getFingerprintsListFunc,
	approveFingerprintFunction approveFingerprintFunc,
	revokeFingerprintFunction revokeFingerprintFunc,
) error {
	client := m.(*fivetran.Client)
	// Verify resource exists
	var id = d.Get(resourceType.String() + "_id").(string)
	resourceId, err := verifyFingerprint(ctx, client, id, resourceType)
	if err != nil {
		return err
	}
	// Sync fingerprints
	err = syncFingerprints(ctx, client, d.Get("fingerprint").(*schema.Set).List(), id,
		getFingerprintsFunction,
		approveFingerprintFunction,
		revokeFingerprintFunction,
	)
	if err != nil {
		return err
	}
	d.SetId(resourceId)
	return nil
}

func resourceFingerprintsRead(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType) error {
	switch resourceType {
	case Connector:
		return resourceFingerprintsReadBase(ctx, d, m, resourceType, getConnectorFingerprintsFunction)
	case Destination:
		return resourceFingerprintsReadBase(ctx, d, m, resourceType, getDestinationFingerprintsFunction)
	}
	return fmt.Errorf("unknown resource type " + resourceType.String())
}

func resourceFingerprintsReadBase(ctx context.Context, d *schema.ResourceData, m interface{},
	resourceType ResourceType, fetchFunc getFingerprintsListFunc) error {
	id := d.Get("id").(string)
	response, err := fetchFingerprints(ctx, m.(*fivetran.Client), id, fetchFunc)
	if err != nil {
		return fmt.Errorf("%v; code: %v; message: %v", err, response.Code, response.Message)
	}
	msi := make(map[string]interface{})
	msi[resourceType.String()+"_id"] = id
	msi["fingerprint"] = flattenFingerprints(response)
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func resourceFingerprintsUpdate(ctx context.Context, d *schema.ResourceData, m interface{},
	resourceType ResourceType) error {
	switch resourceType {
	case Connector:
		err := resourceFingerprintsUpdateBase(ctx, d, m, resourceType,
			getConnectorFingerprintsFunction,
			approveConnectorFingerprintFunc,
			revokeConnectorFingerprintFunc)
		if err != nil {
			return err
		}
		return resourceFingerprintsRead(ctx, d, m, Connector)
	case Destination:
		err := resourceFingerprintsUpdateBase(ctx, d, m, resourceType,
			getDestinationFingerprintsFunction,
			approveDestinationFingerprintFunc,
			revokeDestinationFingerprintFunc)
		if err != nil {
			return err
		}
		return resourceFingerprintsRead(ctx, d, m, Connector)
	}
	return fmt.Errorf("unknown resource type " + resourceType.String())
}

func resourceFingerprintsUpdateBase(ctx context.Context, d *schema.ResourceData, m interface{},
	resourceType ResourceType,
	getFingerprintsFunction getFingerprintsListFunc,
	approveFingerprintFunction approveFingerprintFunc,
	revokeFingerprintFunction revokeFingerprintFunc,
) error {
	client := m.(*fivetran.Client)
	connectorId := d.Get("connector_id").(string)
	if d.HasChange("fingerprint") {
		// Sync fingerprints
		err := syncFingerprints(ctx, client, d.Get("fingerprint").(*schema.Set).List(), connectorId,
			getFingerprintsFunction,
			approveFingerprintFunction,
			revokeFingerprintFunction,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceFingerprintsDelete(ctx context.Context, d *schema.ResourceData, m interface{}, resourceType ResourceType) error {
	switch resourceType {
	case Connector:
		return resourceFingerprintsDeleteBase(ctx, d, m, resourceType,
			getConnectorFingerprintsFunction,
			approveConnectorFingerprintFunc,
			revokeConnectorFingerprintFunc)
	case Destination:
		return resourceFingerprintsDeleteBase(ctx, d, m, resourceType,
			getDestinationFingerprintsFunction,
			approveDestinationFingerprintFunc,
			revokeDestinationFingerprintFunc)
	}
	return fmt.Errorf("unknown resource type " + resourceType.String())
}

func resourceFingerprintsDeleteBase(ctx context.Context, d *schema.ResourceData, m interface{},
	resourceType ResourceType,
	getFingerprintsFunction getFingerprintsListFunc,
	approveFingerprintFunction approveFingerprintFunc,
	revokeFingerprintFunction revokeFingerprintFunc,
) error {
	client := m.(*fivetran.Client)
	id := d.Get(resourceType.String() + "_id").(string)

	// Sync with empty local fingerprints list leads to cleanup
	return syncFingerprints(ctx, client, []interface{}{}, id,
		getFingerprintsFunction,
		approveFingerprintFunction,
		revokeFingerprintFunction)
}

func fetchFingerprints(
	ctx context.Context,
	client *fivetran.Client,
	id string,
	fetchFunc getFingerprintsListFunc) (fingerprints.FingerprintsListResponse, error) {
	var result fingerprints.FingerprintsListResponse
	result, err := fetchFunc(ctx, client, id, "")
	cursor := result.Data.NextCursor
	for err == nil && cursor != "" {
		if innerResult, err := fetchFunc(ctx, client, id, cursor); err == nil {
			cursor = innerResult.Data.NextCursor
			result.Data.Items = append(result.Data.Items, innerResult.Data.Items...)
		}
	}
	return result, err
}

func syncFingerprints(
	ctx context.Context,
	client *fivetran.Client,
	local []interface{},
	id string,
	getFingerprintsFunction getFingerprintsListFunc,
	approveFingerprintFunction approveFingerprintFunc,
	revrevokeFingerprintFunction revokeFingerprintFunc) error {
	response, err := fetchFingerprints(ctx, client, id, getFingerprintsFunction)
	if err == nil {
		upstream := make([]string, 0)
		for k := range mapFingerprintsFromResponse(response) {
			upstream = append(upstream, k)
		}
		localFingerprints := mapFingerprintsFromResourceData(local)
		local := make([]string, 0)
		for k := range localFingerprints {
			local = append(local, k)
		}
		revoke, approve := observeDiffFingerprints(upstream, local)

		for _, r := range revoke {
			resp, err := revrevokeFingerprintFunction(ctx, client, id, r)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}

		for _, a := range approve {
			resp, err := approveFingerprintFunction(ctx, client, id, a, localFingerprints[a].(map[string]interface{})["public_key"].(string))
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
	}
	return err
}

func verifyFingerprint(ctx context.Context, client *fivetran.Client, id string, resourceType ResourceType) (string, error) {
	if resourceType == Connector {
		connectorResponse, err := client.NewConnectorDetails().ConnectorID(id).Do(ctx)
		if err != nil {
			return "", err
		}
		return connectorResponse.Data.ID, nil
	} else if resourceType == Destination {
		destinationResponse, err := client.NewDestinationDetails().DestinationID(id).Do(ctx)
		if err != nil {
			return "", err
		}
		return destinationResponse.Data.ID, nil
	} else {
		return "", fmt.Errorf("unknown resource type %v", resourceType.String())
	}
}

// Function accepts two sets of item hashes: upstream items list and local items list
// Returns list of items to revoke from upstream and list of items to approve in upstream
func observeDiffFingerprints(upstream, local []string) (revoke, approve []string) {
	return []string{}, []string{}
}

func mapFingerprintsFromResponse(resp fingerprints.FingerprintsListResponse) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fp := range resp.Data.Items {
		result[fp.Hash] = flattenFingerprint(fp)
	}
	return result
}

func mapFingerprintsFromResourceData(fingerprints []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fp := range fingerprints {
		result[fp.(map[string]interface{})["hash"].(string)] = fp
	}
	return result
}

func flattenFingerprints(resp fingerprints.FingerprintsListResponse) []interface{} {
	result := make([]interface{}, 0)
	for _, fp := range resp.Data.Items {
		result = append(result, flattenFingerprint(fp))
	}
	return result
}

func flattenFingerprint(resp fingerprints.FingerprintDetails) map[string]interface{} {
	f := make(map[string]interface{})
	f["hash"] = resp.Hash
	f["public_key"] = resp.PublicKey
	f["validated_by"] = resp.ValidatedBy
	f["validated_date"] = resp.ValidatedDate
	return f
}

func fingerprintItemHash(v interface{}) int {
	h := fnv.New32a()
	var hashKey = v.(map[string]interface{})["hash"].(string)
	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

// Connectors
func getConnectorFingerprintsFunction(ctx context.Context, client *fivetran.Client, id, cursor string) (fingerprints.FingerprintsListResponse, error) {
	svc := client.NewConnectorFingerprintsList().ConnectorID(id)
	if cursor == "" {
		svc.Cursor(cursor)
	}
	return svc.Do(ctx)
}

func approveConnectorFingerprintFunc(ctx context.Context, client *fivetran.Client, id, hash, publicKey string) (fingerprints.FingerprintResponse, error) {
	return client.NewCertificateConnectorFingerprintApprove().
		ConnectorID(id).
		Hash(hash).
		PublicKey(publicKey).
		Do(ctx)
}

func revokeConnectorFingerprintFunc(ctx context.Context, client *fivetran.Client, id, hash string) (common.CommonResponse, error) {
	return client.NewConnectorFingerprintRevoke().
		ConnectorID(id).
		Hash(hash).
		Do(ctx)
}

// Destinations
func getDestinationFingerprintsFunction(ctx context.Context, client *fivetran.Client, id, cursor string) (fingerprints.FingerprintsListResponse, error) {
	svc := client.NewDestinationFingerprintsList().DestinationID(id)
	if cursor == "" {
		svc.Cursor(cursor)
	}
	return svc.Do(ctx)
}

func approveDestinationFingerprintFunc(ctx context.Context, client *fivetran.Client, id, hash, publicKey string) (fingerprints.FingerprintResponse, error) {
	return client.NewCertificateDestinationFingerprintApprove().
		DestinationID(id).
		Hash(hash).
		PublicKey(publicKey).
		Do(ctx)
}

func revokeDestinationFingerprintFunc(ctx context.Context, client *fivetran.Client, id, hash string) (common.CommonResponse, error) {
	return client.NewDestinationFingerprintRevoke().
		DestinationID(id).
		Hash(hash).
		Do(ctx)
}
