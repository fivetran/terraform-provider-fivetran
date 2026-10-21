package core

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
)

// The connector holds set-typed fields in a Java HashSet, so the API returns them in an arbitrary
// order that can change between two reads of an unchanged connection. These tests cover how the
// projection (Read/Create/Update state), the PATCH diff and the equality helpers handle that.

const setTestService = "facebook_ads"

var attributionWindows = []interface{}{"CLICK_1D", "CLICK_7D", "CLICK_28D", "VIEW_1D"}

// permutations returns every ordering of items.
func permutations(items []interface{}) [][]interface{} {
	if len(items) <= 1 {
		return [][]interface{}{append([]interface{}{}, items...)}
	}
	var result [][]interface{}
	for i := range items {
		rest := make([]interface{}, 0, len(items)-1)
		rest = append(rest, items[:i]...)
		rest = append(rest, items[i+1:]...)
		for _, p := range permutations(rest) {
			result = append(result, append([]interface{}{items[i]}, p...))
		}
	}
	return result
}

func customTable(name string, windows []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"table_name":          name,
		"aggregation":         "Day",
		"attribution_windows": windows,
		"fields":              []interface{}{"spend", "impressions", "clicks"},
		"breakdowns":          []interface{}{"age", "gender"},
		"action_breakdowns":   []interface{}{"action_type", "action_device"},
	}
}

func customTablesConfig(tables ...map[string]interface{}) map[string]interface{} {
	items := make([]interface{}, len(tables))
	for i, table := range tables {
		items[i] = table
	}
	return map[string]interface{}{"custom_tables": items}
}

func facebookAdsSlot() *metadata.Property {
	return makeSlot(map[string]*metadata.Property{
		"accounts": {Type: "array", Items: &metadata.Property{Type: "string"}},
		"custom_tables": {
			Type: "array",
			Items: &metadata.Property{
				Type: "object",
				Properties: map[string]*metadata.Property{
					"table_name":          {Type: "string"},
					"aggregation":         {Type: "string"},
					"attribution_windows": {Type: "array", Items: &metadata.Property{Type: "string"}},
					"fields":              {Type: "array", Items: &metadata.Property{Type: "string"}},
					"breakdowns":          {Type: "array", Items: &metadata.Property{Type: "string"}},
					"action_breakdowns":   {Type: "array", Items: &metadata.Property{Type: "string"}},
				},
			},
		},
		"ordered_list": {Type: "array", Items: &metadata.Property{Type: "string"}},
	})
}

func assertDeepEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

// --- alignSetOrder ---

func TestAlignSetOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		remote []interface{}
		prior  []interface{}
		want   []interface{}
	}{
		{
			name:   "reorder only keeps prior order",
			remote: []interface{}{"VIEW_1D", "CLICK_28D", "CLICK_1D", "CLICK_7D"},
			prior:  []interface{}{"CLICK_1D", "CLICK_7D", "CLICK_28D", "VIEW_1D"},
			want:   []interface{}{"CLICK_1D", "CLICK_7D", "CLICK_28D", "VIEW_1D"},
		},
		{
			name:   "element added upstream is appended after prior elements",
			remote: []interface{}{"VIEW_1D", "CLICK_28D", "CLICK_1D"},
			prior:  []interface{}{"CLICK_1D", "VIEW_1D"},
			want:   []interface{}{"CLICK_1D", "VIEW_1D", "CLICK_28D"},
		},
		{
			name:   "element removed upstream is dropped, rest keeps prior order",
			remote: []interface{}{"VIEW_1D", "CLICK_1D"},
			prior:  []interface{}{"CLICK_1D", "CLICK_28D", "VIEW_1D"},
			want:   []interface{}{"CLICK_1D", "VIEW_1D"},
		},
		{
			name:   "added and removed at once",
			remote: []interface{}{"c", "x", "a"},
			prior:  []interface{}{"a", "b", "c"},
			want:   []interface{}{"a", "c", "x"},
		},
		{
			name:   "new elements keep their API order",
			remote: []interface{}{"z", "a", "y"},
			prior:  []interface{}{"a"},
			want:   []interface{}{"a", "z", "y"},
		},
		{
			name:   "duplicates are matched by count",
			remote: []interface{}{"b", "a", "a"},
			prior:  []interface{}{"a", "b", "a", "a"},
			want:   []interface{}{"a", "b", "a"},
		},
		{
			name:   "empty remote clears the list",
			remote: []interface{}{},
			prior:  []interface{}{"a"},
			want:   []interface{}{},
		},
		{
			name:   "empty prior takes remote order",
			remote: []interface{}{"b", "a"},
			prior:  []interface{}{},
			want:   []interface{}{"b", "a"},
		},
		{
			name:   "completely different content takes remote order",
			remote: []interface{}{"y", "x"},
			prior:  []interface{}{"a", "b"},
			want:   []interface{}{"y", "x"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertDeepEqual(t, alignSetOrder(tt.remote, tt.prior), tt.want)
		})
	}
}

func TestAlignSetOrder_NumbersMatchAcrossTypesAndKeepPriorType(t *testing.T) {
	t.Parallel()
	// State values come back as int64, the API read-back decodes JSON numbers as float64.
	got := alignSetOrder([]interface{}{float64(3), float64(1), float64(2)}, []interface{}{int64(1), int64(2), int64(3)})
	assertDeepEqual(t, got, []interface{}{int64(1), int64(2), int64(3)})
}

func TestAlignSetOrder_ObjectElements(t *testing.T) {
	t.Parallel()
	a := map[string]interface{}{"id": "a", "n": int64(1)}
	b := map[string]interface{}{"id": "b", "n": int64(2)}
	got := alignSetOrder(
		[]interface{}{map[string]interface{}{"id": "b", "n": float64(2)}, map[string]interface{}{"id": "a", "n": float64(1)}},
		[]interface{}{a, b},
	)
	assertDeepEqual(t, got, []interface{}{a, b})
}

// --- ProjectDynamicWithService ---

func TestProjectWithService_TopLevelSetReorderKeepsMaskOrder(t *testing.T) {
	t.Parallel()
	mask := map[string]interface{}{"accounts": []interface{}{"act_1", "act_2", "act_3"}}
	remote := map[string]interface{}{"accounts": []interface{}{"act_3", "act_1", "act_2"}}

	got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, got["accounts"], mask["accounts"])
}

func TestProjectWithService_EveryRemoteOrderConvergesToMask(t *testing.T) {
	t.Parallel()
	// The API can return any ordering on any read, so every one of them must project to the
	// prior (mask) value. Otherwise state churns and the next plan is not empty.
	mask := customTablesConfig(customTable("t1", attributionWindows))
	for _, order := range permutations(attributionWindows) {
		remoteTable := customTable("t1", order)
		remoteTable["fields"] = []interface{}{"clicks", "spend", "impressions"}
		remoteTable["breakdowns"] = []interface{}{"gender", "age"}
		remoteTable["action_breakdowns"] = []interface{}{"action_device", "action_type"}
		remote := customTablesConfig(remoteTable)

		got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
		if !reflect.DeepEqual(got, mask) {
			t.Fatalf("remote order %v: got %#v, want mask %#v", order, got, mask)
		}
	}
}

func TestProjectWithService_MultipleCustomTablesAlignedIndependently(t *testing.T) {
	t.Parallel()
	mask := customTablesConfig(
		customTable("t1", []interface{}{"CLICK_1D", "VIEW_1D"}),
		customTable("t2", []interface{}{"CLICK_7D", "CLICK_28D"}),
	)
	remote := customTablesConfig(
		customTable("t1", []interface{}{"VIEW_1D", "CLICK_1D"}),
		customTable("t2", []interface{}{"CLICK_28D", "CLICK_7D", "VIEW_1D"}),
	)

	got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
	tables := got["custom_tables"].([]interface{})
	assertDeepEqual(t, tables[0].(map[string]interface{})["attribution_windows"], []interface{}{"CLICK_1D", "VIEW_1D"})
	// Real drift in t2 shows only the added element, after the prior ones.
	assertDeepEqual(t, tables[1].(map[string]interface{})["attribution_windows"], []interface{}{"CLICK_7D", "CLICK_28D", "VIEW_1D"})
}

func TestProjectWithService_ScalarDriftInsideCustomTableStillSurfaces(t *testing.T) {
	t.Parallel()
	mask := customTablesConfig(customTable("t1", attributionWindows))
	remoteTable := customTable("t1", []interface{}{"VIEW_1D", "CLICK_28D", "CLICK_7D", "CLICK_1D"})
	remoteTable["aggregation"] = "Week"
	remote := customTablesConfig(remoteTable)

	got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
	table := got["custom_tables"].([]interface{})[0].(map[string]interface{})
	assertDeepEqual(t, table["aggregation"], "Week")
	assertDeepEqual(t, table["attribution_windows"], attributionWindows)
}

func TestProjectWithService_UpstreamOnlyKeysInCustomTableDropped(t *testing.T) {
	t.Parallel()
	mask := customTablesConfig(customTable("t1", attributionWindows))
	remoteTable := customTable("t1", []interface{}{"VIEW_1D", "CLICK_1D", "CLICK_7D", "CLICK_28D"})
	remoteTable["use_unified_attribution_setting"] = true
	remote := customTablesConfig(remoteTable)

	got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, got, mask)
}

func TestProjectWithService_OrderedListKeepsRemoteOrder(t *testing.T) {
	t.Parallel()
	// Not every list is a set: for ordered lists a reorder is real drift.
	mask := map[string]interface{}{"ordered_list": []interface{}{"a", "b"}}
	remote := map[string]interface{}{"ordered_list": []interface{}{"b", "a"}}

	got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, got["ordered_list"], []interface{}{"b", "a"})
}

func TestProjectWithService_OtherServiceKeepsRemoteOrder(t *testing.T) {
	t.Parallel()
	// The same field name is not a set for a service that is not listed.
	mask := map[string]interface{}{"accounts": []interface{}{"a", "b"}}
	remote := map[string]interface{}{"accounts": []interface{}{"b", "a"}}

	assertDeepEqual(t, ProjectDynamicWithService(remote, mask, nil, "google_ads")["accounts"], []interface{}{"b", "a"})
	assertDeepEqual(t, ProjectDynamic(remote, mask, nil)["accounts"], []interface{}{"b", "a"})
}

func TestProjectWithService_SetPathOnlyMatchesAtThatPath(t *testing.T) {
	t.Parallel()
	// "custom_tables.fields" is a set, but a "fields" list under another object is not.
	mask := map[string]interface{}{"other": map[string]interface{}{"attribution_windows": []interface{}{"a", "b"}}}
	remote := map[string]interface{}{"other": map[string]interface{}{"attribution_windows": []interface{}{"b", "a"}}}

	got := ProjectDynamicWithService(remote, mask, nil, setTestService)
	assertDeepEqual(t, got["other"].(map[string]interface{})["attribution_windows"], []interface{}{"b", "a"})
}

func TestProjectWithService_CustomTablesCountChangedKeepsRemote(t *testing.T) {
	t.Parallel()
	// custom_tables is an ordered list (a Java List), paired by index. When a table is added
	// upstream the pairing is unreliable, so the whole remote list is kept to surface drift.
	mask := customTablesConfig(customTable("t1", attributionWindows))
	remote := customTablesConfig(
		customTable("t1", attributionWindows),
		customTable("t2", attributionWindows),
	)

	got := ProjectDynamicWithService(remote, mask, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, got["custom_tables"], remote["custom_tables"])
}

func TestProjectWithService_SetFieldMissingFromRemoteKeepsMask(t *testing.T) {
	t.Parallel()
	mask := map[string]interface{}{"accounts": []interface{}{"a", "b"}}

	got := ProjectDynamicWithService(map[string]interface{}{}, mask, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, got["accounts"], []interface{}{"a", "b"})
}

func TestProjectWithService_SetFieldNullInRemoteSurfacesNull(t *testing.T) {
	t.Parallel()
	mask := map[string]interface{}{"accounts": []interface{}{"a"}}

	got := ProjectDynamicWithService(map[string]interface{}{"accounts": nil}, mask, facebookAdsSlot(), setTestService)
	if v, ok := got["accounts"]; !ok || v != nil {
		t.Errorf("expected accounts to be nil, got %#v", v)
	}
}

// --- PrepareConfigPatchDynamicWithService ---

func TestPatchWithService_NestedSetReorderNotSent(t *testing.T) {
	t.Parallel()
	// Before the fix only top-level keys were compared with set semantics, so a reorder inside
	// custom_tables still resent the whole custom_tables list.
	plan := customTablesConfig(customTable("t1", []interface{}{"CLICK_1D", "CLICK_7D", "CLICK_28D", "VIEW_1D"}))
	stateTable := customTable("t1", []interface{}{"VIEW_1D", "CLICK_28D", "CLICK_7D", "CLICK_1D"})
	stateTable["fields"] = []interface{}{"clicks", "impressions", "spend"}
	state := customTablesConfig(stateTable)

	patch := PrepareConfigPatchDynamicWithService(plan, state, facebookAdsSlot(), setTestService)
	if len(patch) != 0 {
		t.Errorf("reorder-only change must not be sent, got patch %#v", patch)
	}
}

func TestPatchWithService_NestedSetChangeSendsCustomTablesInPlanOrder(t *testing.T) {
	t.Parallel()
	plan := customTablesConfig(customTable("t1", []interface{}{"CLICK_1D", "VIEW_1D", "CLICK_28D"}))
	state := customTablesConfig(customTable("t1", []interface{}{"VIEW_1D", "CLICK_1D"}))

	patch := PrepareConfigPatchDynamicWithService(plan, state, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, patch["custom_tables"], plan["custom_tables"])
}

func TestPatchWithService_TopLevelSetReorderNotSent(t *testing.T) {
	t.Parallel()
	patch := PrepareConfigPatchDynamicWithService(
		map[string]interface{}{"accounts": []interface{}{"a", "b", "c"}},
		map[string]interface{}{"accounts": []interface{}{"c", "a", "b"}},
		facebookAdsSlot(), setTestService,
	)
	if len(patch) != 0 {
		t.Errorf("reorder-only change must not be sent, got patch %#v", patch)
	}
}

func TestPatchWithService_OrderedListReorderSent(t *testing.T) {
	t.Parallel()
	patch := PrepareConfigPatchDynamicWithService(
		map[string]interface{}{"ordered_list": []interface{}{"a", "b"}},
		map[string]interface{}{"ordered_list": []interface{}{"b", "a"}},
		facebookAdsSlot(), setTestService,
	)
	assertDeepEqual(t, patch["ordered_list"], []interface{}{"a", "b"})
}

func TestPatchWithService_NonSetSiblingChangeStillSent(t *testing.T) {
	t.Parallel()
	planTable := customTable("t1", attributionWindows)
	planTable["aggregation"] = "Week"
	plan := customTablesConfig(planTable)
	state := customTablesConfig(customTable("t1", []interface{}{"VIEW_1D", "CLICK_28D", "CLICK_7D", "CLICK_1D"}))

	patch := PrepareConfigPatchDynamicWithService(plan, state, facebookAdsSlot(), setTestService)
	assertDeepEqual(t, patch["custom_tables"], plan["custom_tables"])
}

func TestPatchWithService_CustomTablesReorderedIsSent(t *testing.T) {
	t.Parallel()
	// custom_tables itself is ordered, so swapping two tables is a change.
	t1 := customTable("t1", attributionWindows)
	t2 := customTable("t2", attributionWindows)
	patch := PrepareConfigPatchDynamicWithService(customTablesConfig(t2, t1), customTablesConfig(t1, t2), facebookAdsSlot(), setTestService)
	if _, ok := patch["custom_tables"]; !ok {
		t.Error("reordering custom_tables must be sent")
	}
}

func TestPatchWithService_DuplicatesCountAsChange(t *testing.T) {
	t.Parallel()
	patch := PrepareConfigPatchDynamicWithService(
		map[string]interface{}{"accounts": []interface{}{"a", "a", "b"}},
		map[string]interface{}{"accounts": []interface{}{"a", "b", "b"}},
		facebookAdsSlot(), setTestService,
	)
	if _, ok := patch["accounts"]; !ok {
		t.Error("lists with different element counts must be sent")
	}
}

func TestPatchWithoutService_ReorderSent(t *testing.T) {
	t.Parallel()
	patch := PrepareConfigPatchDynamic(
		map[string]interface{}{"accounts": []interface{}{"a", "b"}},
		map[string]interface{}{"accounts": []interface{}{"b", "a"}},
		facebookAdsSlot(),
	)
	assertDeepEqual(t, patch["accounts"], []interface{}{"a", "b"})
}

// --- DynamicValuesEqualAtPath ---

func TestDynamicValuesEqualAtPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		left, right interface{}
		path        string
		want        bool
	}{
		{"top-level set reordered", []interface{}{"a", "b"}, []interface{}{"b", "a"}, "accounts", true},
		{"top-level set changed", []interface{}{"a", "b"}, []interface{}{"a", "c"}, "accounts", false},
		{"ordered list reordered", []interface{}{"a", "b"}, []interface{}{"b", "a"}, "ordered_list", false},
		{"set at root path not recognized", []interface{}{"a", "b"}, []interface{}{"b", "a"}, "", false},
		{
			"custom_tables value with nested set reordered",
			[]interface{}{map[string]interface{}{"attribution_windows": []interface{}{"a", "b"}}},
			[]interface{}{map[string]interface{}{"attribution_windows": []interface{}{"b", "a"}}},
			"custom_tables", true,
		},
		{"set of numbers across types", []interface{}{int64(1), int64(2)}, []interface{}{float64(2), float64(1)}, "accounts", true},
		{"nil and nil", nil, nil, "accounts", true},
		{"nil and empty list", nil, []interface{}{}, "accounts", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DynamicValuesEqualAtPath(tt.left, tt.right, setTestService, tt.path); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSetTypedField_FacebookAdsFields(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"accounts",
		"custom_prebuilt_report_table_names",
		"conflict_table_names",
		"custom_tables.fields",
		"custom_tables.breakdowns",
		"custom_tables.action_breakdowns",
		"custom_tables.attribution_windows",
	} {
		if !IsSetTypedField(setTestService, path) {
			t.Errorf("%s should be set-typed for %s", path, setTestService)
		}
	}
	for _, path := range []string{"custom_tables", "custom_tables.table_name", "sync_mode"} {
		if IsSetTypedField(setTestService, path) {
			t.Errorf("%s should not be set-typed", path)
		}
	}
	if IsSetTypedField("", "accounts") {
		t.Error("an empty service has no set-typed fields")
	}
}

// --- setElementKey / multisetEqual ---

func TestMultisetEqual_ElementKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b []interface{}
		want bool
	}{
		{"same strings reordered", []interface{}{"a", "b", "c"}, []interface{}{"c", "a", "b"}, true},
		{"different lengths", []interface{}{"a", "b"}, []interface{}{"a", "b", "b"}, false},
		{"same length, different duplicates", []interface{}{"a", "a", "b"}, []interface{}{"a", "b", "b"}, false},
		{"string is not a number", []interface{}{"1"}, []interface{}{float64(1)}, false},
		{"string is not a bool", []interface{}{"true"}, []interface{}{true}, false},
		{"string is not null", []interface{}{"null"}, []interface{}{nil}, false},
		{"int64 equals float64 of the same value", []interface{}{int64(2), int64(1)}, []interface{}{float64(1), float64(2)}, true},
		{"fractional number differs", []interface{}{float64(1.5)}, []interface{}{int64(1)}, false},
		{
			"objects with the same content reordered",
			[]interface{}{map[string]interface{}{"x": "1", "y": int64(2)}, map[string]interface{}{"x": "3"}},
			[]interface{}{map[string]interface{}{"x": "3"}, map[string]interface{}{"y": float64(2), "x": "1"}},
			true,
		},
		{
			"objects with different content",
			[]interface{}{map[string]interface{}{"x": "1"}},
			[]interface{}{map[string]interface{}{"x": "2"}},
			false,
		},
		{"nested lists keep their own order", []interface{}{[]interface{}{"a", "b"}}, []interface{}{[]interface{}{"b", "a"}}, false},
		{"unknown never equals an empty object", []interface{}{DynamicUnknownValue{}}, []interface{}{map[string]interface{}{}}, false},
		{"unknown never equals unknown", []interface{}{DynamicUnknownValue{}}, []interface{}{DynamicUnknownValue{}}, false},
		{"both empty", []interface{}{}, []interface{}{}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := multisetEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("multisetEqual(%#v, %#v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
			if got := multisetEqual(tt.b, tt.a); got != tt.want {
				t.Errorf("multisetEqual is not symmetric for %#v, %#v", tt.a, tt.b)
			}
		})
	}
}

func TestAlignSetOrder_TypedLookalikesAreNotMatched(t *testing.T) {
	t.Parallel()
	// "1" and 1 encode differently, so the prior string is dropped and the remote number kept.
	got := alignSetOrder([]interface{}{float64(1), "a"}, []interface{}{"a", "1"})
	assertDeepEqual(t, got, []interface{}{"a", float64(1)})
}

func TestAlignSetOrder_LargeListKeepsPriorOrder(t *testing.T) {
	t.Parallel()
	prior := make([]interface{}, 500)
	remote := make([]interface{}, 500)
	for i := range prior {
		prior[i] = "item_" + strconv.Itoa(i)
		remote[(i*7)%500] = prior[i]
	}
	assertDeepEqual(t, alignSetOrder(remote, prior), prior)
	if !multisetEqual(remote, prior) {
		t.Error("a permutation must be multiset-equal")
	}
}
