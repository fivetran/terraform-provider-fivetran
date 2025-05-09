package model

import (
	"context"

	"github.com/fivetran/go-fivetran/certificates"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type CertificateConnector struct {
	Id          types.String `tfsdk:"id"`
	ConnectorId types.String `tfsdk:"connector_id"`
	Certificate types.Set    `tfsdk:"certificate"`
}

type CertificatesConnector struct {
	Id           types.String `tfsdk:"id"`
	ConnectorId  types.String `tfsdk:"connector_id"`
	Certificates types.Set    `tfsdk:"certificates"`
}

type CertificateConnection struct {
	Id           types.String `tfsdk:"id"`
	Certificate  types.Set    `tfsdk:"certificate"`
}

type CertificatesConnection struct {
	Id            types.String `tfsdk:"id"`
	Certificates  types.Set    `tfsdk:"certificates"`
}

type CertificateDestination struct {
	Id            types.String `tfsdk:"id"`
	DestinationId types.String `tfsdk:"destination_id"`
	Certificate   types.Set    `tfsdk:"certificate"`
}

type CertificatesDestination struct {
	Id            types.String `tfsdk:"id"`
	DestinationId types.String `tfsdk:"destination_id"`
	Certificates  types.Set    `tfsdk:"certificates"`
}

var (
	elementCertificateType = map[string]attr.Type{
		"hash":           types.StringType,
		"public_key":     types.StringType,
		"name":           types.StringType,
		"type":           types.StringType,
		"sha1":           types.StringType,
		"sha256":         types.StringType,
		"validated_by":   types.StringType,
		"validated_date": types.StringType,
		"encoded_cert":   types.StringType,
	}

	elementDatasourceCertificateType = map[string]attr.Type{
		"hash":           types.StringType,
		"public_key":     types.StringType,
		"name":           types.StringType,
		"type":           types.StringType,
		"sha1":           types.StringType,
		"sha256":         types.StringType,
		"validated_by":   types.StringType,
		"validated_date": types.StringType,
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

func (d *CertificateConnector) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
	d.Id = d.ConnectorId
	d.Certificate = readCertificateItemsFromResponse(resp, d.getEncodedCertsMap(), true)
}

func (d *CertificatesConnector) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
	d.ConnectorId = d.Id
	emptyMap := make(map[string]string)
	d.Certificates = readCertificateItemsFromResponse(resp, emptyMap, false)
}

func (d *CertificateConnection) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
	d.Certificate = readCertificateItemsFromResponse(resp, d.getEncodedCertsMap(), true)
}

func (d *CertificatesConnection) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
	emptyMap := make(map[string]string)
	d.Certificates = readCertificateItemsFromResponse(resp, emptyMap, false)
}

func (d *CertificateDestination) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
	d.Id = d.DestinationId
	d.Certificate = readCertificateItemsFromResponse(resp, d.getEncodedCertsMap(), true)
}

func (d *CertificatesDestination) ReadFromResponse(ctx context.Context, resp certificates.CertificatesListResponse) {
	d.DestinationId = d.Id
	emptyMap := make(map[string]string)
	d.Certificates = readCertificateItemsFromResponse(resp, emptyMap, false)
}

func (d *CertificateConnector) getEncodedCertsMap() map[string]string {
	return getEncodedCertsMapImpl(d.Certificate.Elements())
}

func (d *CertificateConnection) getEncodedCertsMap() map[string]string {
	return getEncodedCertsMapImpl(d.Certificate.Elements())
}

func (d *CertificateDestination) getEncodedCertsMap() map[string]string {
	return getEncodedCertsMapImpl(d.Certificate.Elements())
}

func getEncodedCertsMapImpl(elements []attr.Value) map[string]string {
	result := map[string]string{}

	for _, item := range elements {
		if element, ok := item.(basetypes.ObjectValue); ok {
			encodedCertValue := element.Attributes()["encoded_cert"].(basetypes.StringValue)
			hashValue := element.Attributes()["hash"].(basetypes.StringValue)

			if !hashValue.IsNull() && !hashValue.IsUnknown() && !encodedCertValue.IsNull() && !encodedCertValue.IsUnknown() {
				result[hashValue.ValueString()] = encodedCertValue.ValueString()
			}
		}
	}

	return result
}
