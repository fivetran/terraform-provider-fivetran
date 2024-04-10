package model

import (
	"context"

	"github.com/fivetran/go-fivetran/fingerprints"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FingerprintConnector struct {
	Id          types.String `tfsdk:"id"`
	ConnectorId types.String `tfsdk:"connector_id"`
	Fingerprint types.Set    `tfsdk:"fingerprint"`
}

type FingerprintsConnector struct {
	Id           types.String `tfsdk:"id"`
	ConnectorId  types.String `tfsdk:"connector_id"`
	Fingerprints types.Set    `tfsdk:"fingerprints"`
}

type FingerprintDestination struct {
	Id            types.String `tfsdk:"id"`
	DestinationId types.String `tfsdk:"destination_id"`
	Fingerprint   types.Set    `tfsdk:"fingerprint"`
}

type FingerprintsDestination struct {
	Id            types.String `tfsdk:"id"`
	DestinationId types.String `tfsdk:"destination_id"`
	Fingerprints  types.Set    `tfsdk:"fingerprints"`
}

var (
	elementFingerprintType = map[string]attr.Type{
		"hash":           types.StringType,
		"public_key":     types.StringType,
		"validated_by":   types.StringType,
		"validated_date": types.StringType,
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
	d.ConnectorId = d.Id
	d.Fingerprints = readFingerprintItemsFromResponse(resp)
}

func (d *FingerprintDestination) ReadFromResponse(ctx context.Context, resp fingerprints.FingerprintsListResponse) {
	d.Id = d.DestinationId
	d.Fingerprint = readFingerprintItemsFromResponse(resp)
}

func (d *FingerprintsDestination) ReadFromResponse(ctx context.Context, resp fingerprints.FingerprintsListResponse) {
	d.DestinationId = d.Id
	d.Fingerprints = readFingerprintItemsFromResponse(resp)
}
