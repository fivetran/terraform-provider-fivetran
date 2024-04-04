package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/certificates"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type CertificateConnector struct {
    Id             types.String `tfsdk:"id"`
    ConnectorId    types.String `tfsdk:"connector_id"` 
    Certificate    types.Set    `tfsdk:"certificate"`
}

type CertificatesConnector struct {
    Id             types.String `tfsdk:"id"`
    ConnectorId    types.String `tfsdk:"connector_id"` 
    Certificates   types.Set    `tfsdk:"certificates"`
}

type CertificateDestination struct {
    Id             types.String `tfsdk:"id"`
    DestinationId  types.String `tfsdk:"destination_id"` 
    Certificate    types.Set    `tfsdk:"certificate"`
}

type CertificatesDestination struct {
    Id             types.String `tfsdk:"id"`
    DestinationId  types.String `tfsdk:"destination_id"` 
    Certificates   types.Set    `tfsdk:"certificates"`
}

var (
    elementCertificateType = map[string]attr.Type{
        "hash":             types.StringType,
        "public_key":       types.StringType,
        "name":             types.StringType,
        "type":             types.StringType,
        "sha1":             types.StringType,
        "sha256":           types.StringType,
        "validated_by":     types.StringType,
        "validated_date":   types.StringType,
        "encoded_cert":     types.StringType,
    }

    elementDatasourceCertificateType = map[string]attr.Type{
        "hash":             types.StringType,
        "public_key":       types.StringType,
        "name":             types.StringType,
        "type":             types.StringType,
        "sha1":             types.StringType,
        "sha256":           types.StringType,
        "validated_by":     types.StringType,
        "validated_date":   types.StringType,
    }
)

func readCertificateItemsFromResponse(resp certificates.CertificatesListResponse, encodedCertMap map[string]string, resource bool) types.Set {
    if resp.Data.Items == nil {
        if resource {
            return types.SetNull(types.ObjectType{AttrTypes: elementCertificateType})    
        } else {
            return types.SetNull(types.ObjectType{AttrTypes: elementDatasourceCertificateType})    
        }
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["hash"] = types.StringValue(v.Hash)
        item["public_key"] = types.StringValue(v.PublicKey)
        item["name"] = types.StringValue(v.Name)
        item["type"] = types.StringValue(v.Type)
        item["sha1"] = types.StringValue(v.Sha1)
        item["sha256"] = types.StringValue(v.Sha256)
        item["validated_by"] = types.StringValue(v.ValidatedBy)
        item["validated_date"] = types.StringValue(v.ValidatedDate)

        if resource {
            encodedCertValue, found := encodedCertMap[v.Hash]
            if found {
                item["encoded_cert"] = types.StringValue(encodedCertValue)
            }            
        }

        var objectValue types.Object
        if resource {
            objectValue, _ = types.ObjectValue(elementCertificateType, item)
        } else {
            objectValue, _ = types.ObjectValue(elementDatasourceCertificateType, item)
        }

        items = append(items, objectValue)
    }

    var result types.Set
    if resource {
        result, _ = types.SetValue(types.ObjectType{AttrTypes: elementCertificateType}, items)
    } else {
        result, _ = types.SetValue(types.ObjectType{AttrTypes: elementDatasourceCertificateType}, items)
    }
    
    return result
}

func (d *CertificateConnector) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse, encodedCertMap map[string]string) {
    d.Id = d.ConnectorId
    d.Certificate = readCertificateItemsFromResponse(resp, encodedCertMap, true)
}

func (d *CertificateDestination) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse, encodedCertMap map[string]string) {
    d.Id = d.DestinationId
    d.Certificate = readCertificateItemsFromResponse(resp, encodedCertMap, true)
}

func (d *CertificatesConnector) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
    d.Id = d.ConnectorId
    emptyMap := make(map[string]string)
    d.Certificates = readCertificateItemsFromResponse(resp, emptyMap, false)
}

func (d *CertificatesDestination) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
    d.Id = d.DestinationId
    emptyMap := make(map[string]string)
    d.Certificates = readCertificateItemsFromResponse(resp, emptyMap, false)
}

func readFromSourceConnectorCommon(ctx context.Context, client *fivetran.Client, id string, service string) (certificates.CertificatesListResponse, error) {
    var respNextCursor string
    var listResponse certificates.CertificatesListResponse
    limit := 1000

    for {
        var err error
        var tmpResp certificates.CertificatesListResponse

        if service == "CertificateConnector" || service == "CertificatesConnector" {
            svc := client.NewConnectorCertificatesList().ConnectorID(id).Limit(limit)
            if respNextCursor != "" {
                svc.Cursor(respNextCursor)
            }
            tmpResp, err = svc.Do(ctx)
        }

        if service == "CertificateDestination" || service == "CertificatesDestination" {
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

func (d *CertificateConnector) ReadFromSource(ctx context.Context, client *fivetran.Client, connectorId string) (certificates.CertificatesListResponse, error) {
    return readFromSourceConnectorCommon(ctx, client, connectorId, "CertificateConnector")
}

func (d *CertificatesConnector) ReadFromSource(ctx context.Context, client *fivetran.Client, connectorId string) (certificates.CertificatesListResponse, error) {
    return readFromSourceConnectorCommon(ctx, client, connectorId, "CertificatesConnector")
}

func (d *CertificateDestination) ReadFromSource(ctx context.Context, client *fivetran.Client, destinationId string) (certificates.CertificatesListResponse, error) {
    return readFromSourceConnectorCommon(ctx, client, destinationId, "CertificateDestination")
}

func (d *CertificatesDestination) ReadFromSource(ctx context.Context, client *fivetran.Client, destinationId string) (certificates.CertificatesListResponse, error) {
    return readFromSourceConnectorCommon(ctx, client, destinationId, "CertificatesDestination")
}