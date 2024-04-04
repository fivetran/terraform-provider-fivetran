package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/fingerprints"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type FingerprintConnector struct {
    Id             types.String `tfsdk:"id"`
    ConnectorId    types.String `tfsdk:"connector_id"` 
    Fingerprint    types.Set    `tfsdk:"fingerprint"`
}

type FingerprintsConnector struct {
    Id             types.String `tfsdk:"id"`
    ConnectorId    types.String `tfsdk:"connector_id"` 
    Fingerprints   types.Set    `tfsdk:"fingerprints"`
}

type FingerprintDestination struct {
    Id             types.String `tfsdk:"id"`
    DestinationId  types.String `tfsdk:"destination_id"` 
    Fingerprint    types.Set    `tfsdk:"fingerprint"`
}

type FingerprintsDestination struct {
    Id             types.String `tfsdk:"id"`
    DestinationId  types.String `tfsdk:"destination_id"` 
    Fingerprints   types.Set    `tfsdk:"fingerprints"`
}

var (
    elementFingerprintType = map[string]attr.Type{
        "hash":             types.StringType,
        "public_key":       types.StringType,
        "validated_by":     types.StringType,
        "validated_date":   types.StringType,
    }
)

func readFingerprintItemsFromResponse(resp fingerprints.FingerprintsListResponse) types.Set {
    if resp.Data.Items == nil {
        return types.SetNull(types.ObjectType{AttrTypes: elementFingerprintType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["hash"] = types.StringValue(v.Hash)
        item["public_key"] = types.StringValue(v.PublicKey)
        item["validated_by"] = types.StringValue(v.ValidatedBy)
        item["validated_date"] = types.StringValue(v.ValidatedDate)

        objectValue, _ := types.ObjectValue(elementFingerprintType, item)
        items = append(items, objectValue)
    }

    result, _ := types.SetValue(types.ObjectType{AttrTypes: elementFingerprintType}, items)
    return result
}

func (d *FingerprintConnector) ReadFromResponse(ctx context.Context, resp fingerprints.FingerprintsListResponse) {
    d.Id = d.ConnectorId
    d.Fingerprint = readFingerprintItemsFromResponse(resp)
}

func (d *FingerprintsConnector) ReadFromResponse(ctx context.Context, resp fingerprints.FingerprintsListResponse) {
    d.Id = d.ConnectorId
    d.Fingerprints = readFingerprintItemsFromResponse(resp)
}

func (d *FingerprintDestination) ReadFromResponse(ctx context.Context, resp fingerprints.FingerprintsListResponse) {
    d.Id = d.DestinationId
    d.Fingerprint = readFingerprintItemsFromResponse(resp)
}

func (d *FingerprintsDestination) ReadFromResponse(ctx context.Context, resp fingerprints.FingerprintsListResponse) {
    d.Id = d.DestinationId
    d.Fingerprints = readFingerprintItemsFromResponse(resp)
}

func readFromSourceFingerprintCommon(ctx context.Context, client *fivetran.Client, id string, service string) (fingerprints.FingerprintsListResponse, error) {
    var respNextCursor string
    var listResponse fingerprints.FingerprintsListResponse
    var err error
    limit := 1000

    for {
        var tmpResp fingerprints.FingerprintsListResponse
        
        if service == "FingerprintConnector" || service == "FingerprintsConnector" {
            svc := client.NewConnectorFingerprintsList().ConnectorID(id).Limit(limit)
            if respNextCursor != "" {
                svc.Cursor(respNextCursor)
            }
            tmpResp, err = svc.Do(ctx)
        }

        if service == "FingerprintDestination" || service == "FingerprintsDestination" {
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

func (d *FingerprintConnector) ReadFromSource(ctx context.Context, client *fivetran.Client, connectorId string) (fingerprints.FingerprintsListResponse, error) {
    return readFromSourceFingerprintCommon(ctx, client, connectorId, "FingerprintConnector")
}

func (d *FingerprintsConnector) ReadFromSource(ctx context.Context, client *fivetran.Client, connectorId string) (fingerprints.FingerprintsListResponse, error) {
    return readFromSourceFingerprintCommon(ctx, client, connectorId, "FingerprintsConnector")
}

func (d *FingerprintDestination) ReadFromSource(ctx context.Context, client *fivetran.Client, destinationId string) (fingerprints.FingerprintsListResponse, error) {
    return readFromSourceFingerprintCommon(ctx, client, destinationId, "FingerprintDestination")
}

func (d *FingerprintsDestination) ReadFromSource(ctx context.Context, client *fivetran.Client, destinationId string) (fingerprints.FingerprintsListResponse, error) {
    return readFromSourceFingerprintCommon(ctx, client, destinationId, "FingerprintsDestination")
}