# implement-tf-field

Add a static API field to the `fivetran_connection_v2` Terraform resource.

## Input

The user will provide a Jira ticket key (e.g. `RD-XXXXXX`). Read the ticket using the Atlassian MCP tool to extract:
- The Java field name (camelCase, e.g. `paused`)
- The DTO class (e.g. `ConnectorResponseV2`)
- The Terraform resource (always `fivetran_connection_v2` for this workflow)
- The API type (string, bool, int, etc.)
- Whether the field is read-only, mutable, or create-only

If the ticket does not contain enough information to determine any of the above, stop and ask the user to clarify rather than guessing.

## Files to change

All paths are relative to the repo root of `terraform-provider-fivetran`:

1. **`fivetran/framework/core/schema/connection_v2.go`** — add schema attribute to `ConnectionV2ResourceAttributes()`
2. **`fivetran/framework/core/model/connection_v2_model.go`** — add field to `ConnectionV2ResourceModel` struct, `ConnectionV2ResourceModelAttrTypes()`, and `readFromResponseData()`
3. **`fivetran/framework/resources/connection_v2.go`** — wire Create/Update request if the field is settable (not read-only)

## Step-by-step

### 1. Read the Jira ticket

Use the Atlassian MCP tool to fetch the ticket. Extract the field name, type, and whether it is read-only/mutable/create-only.

### 2. Check if the field already exists

Read `fivetran/framework/core/schema/connection_v2.go` and check `ConnectionV2ResourceAttributes()`. If the field is already there, stop — nothing to do.

### 3. Check if the SDK has the field

Read `go-fivetran/connections/common_types.go` (in the go-fivetran repo at `../go-fivetran/connections/common_types.go`) and check `DetailsResponseDataCommon`. 

- If the field exists in the SDK → proceed with provider-only changes.
- If the field does NOT exist in the SDK → stop and tell the user that a go-fivetran SDK change is needed first (add the field to `DetailsResponseDataCommon` and open a go-fivetran PR before continuing here).

### 4. Add the schema attribute

In `fivetran/framework/core/schema/connection_v2.go`, add the attribute to `ConnectionV2ResourceAttributes()`.

Map OAS/Java types to Terraform schema types:
- `String` / `string` → `resourceSchema.StringAttribute{}`
- `boolean` / `bool` / `Boolean` → `resourceSchema.BoolAttribute{}`
- `int` / `Integer` / `int64` → `resourceSchema.Int64Attribute{}`

Set `Computed: true` for read-only fields. Set `Optional: true` for mutable or create-only fields. Add a short `Description` matching the field name and type.

Example for a read-only bool field `paused`:
```go
"paused": resourceSchema.BoolAttribute{
    Computed:    true,
    Description: "Specifies whether the connection is paused.",
},
```

Example for a mutable string field `schedule_type`:
```go
"schedule_type": resourceSchema.StringAttribute{
    Optional:    true,
    Computed:    true,
    Description: "The connection sync schedule type.",
},
```

### 5. Add to the model struct

In `fivetran/framework/core/model/connection_v2_model.go`:

a) Add the field to `ConnectionV2ResourceModel`:
```go
Paused types.Bool `tfsdk:"paused"`
```

b) Add the field to `ConnectionV2ResourceModelAttrTypes()`:
```go
"paused": types.BoolType,
```

c) Add the Read mapping in `readFromResponseData()` using the appropriate helper:
- String field: `d.FieldName = stringValueOrNull(data.FieldName)`
- Bool pointer: `d.FieldName = boolPointerValue(data.FieldName)`
- Int pointer: `d.FieldName = intPointerInt64Value(data.FieldName)`
- Time field: `d.FieldName = timeValueOrNull(data.FieldName)`

### 6. Wire Create/Update (if mutable)

Read `fivetran/framework/resources/connection_v2.go` to find where the Create and Update requests are built. If the field is settable (not `Computed`-only), add it to the request builder following the existing pattern.

### 7. Verify

Run the following checks mentally before finishing:
- Schema attribute name matches the `tfsdk` tag in the model struct (snake_case)
- `ConnectionV2ResourceModelAttrTypes()` includes the new field
- `readFromResponseData()` populates the new field
- If the field is settable, Create/Update wires it through

### 8. Report

Tell the user:
- What was added (schema attribute, model field, Read mapping, Create/Update wiring)
- Whether an SDK (go-fivetran) change was needed and if so why it was skipped
- What to verify manually before merging (especially for enum fields or nullable semantics)
- Whether `@TerraformIgnore` should now be removed from the corresponding Java DTO field in the engineering repo

## Guardrails

- Do NOT auto-merge or push anything.
- Do NOT handle `config` or `auth` fields — those are dynamic and handled separately.
- Do NOT handle nested object fields unless they exactly match an existing nested type pattern.
- Do NOT add breaking changes (rename, remove, type change).
- If the field has an enum type in OAS, add a comment listing the valid values but do NOT add validation logic unless the pattern already exists in the resource.
- If unsure about mutability (read-only vs settable), default to `Computed: true` only and note it in the report for human review.
