// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/resources/2023-07-01/resourcegroups"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceGroup2Flatten(t *testing.T) {
	id := commonids.NewResourceGroupID("00000000-0000-0000-0000-000000000001", "example-group")

	tests := []struct {
		name          string
		group         *resourcegroups.ResourceGroup
		wantLocation  string
		wantManagedBy string
		wantTags      map[string]any
	}{
		{
			name: "all fields",
			group: &resourcegroups.ResourceGroup{
				Location:  "West US 2",
				ManagedBy: pointer.To("/subscriptions/example"),
				Tags:      &map[string]string{"environment": "test"},
			},
			wantLocation:  "westus2",
			wantManagedBy: "/subscriptions/example",
			wantTags:      map[string]any{"environment": "test"},
		},
		{
			name: "optional fields absent",
			group: &resourcegroups.ResourceGroup{
				Location: "East US",
			},
			wantLocation: "eastus",
			wantTags:     map[string]any{},
		},
		{
			name:          "nil model preserves existing fields",
			wantLocation:  "existinglocation",
			wantManagedBy: "existing-manager",
			wantTags:      map[string]any{"existing": "tag"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := resourceResourceGroup2().Data(nil)
			for key, value := range map[string]any{
				"location":   "existinglocation",
				"managed_by": "existing-manager",
				"tags":       map[string]any{"existing": "tag"},
			} {
				if err := d.Set(key, value); err != nil {
					t.Fatalf("setting initial %s: %v", key, err)
				}
			}

			if err := resourceResourceGroup2Flatten(d, &id, test.group); err != nil {
				t.Fatalf("flattening resource group: %v", err)
			}

			for key, want := range map[string]any{
				"name":       "example-group",
				"location":   test.wantLocation,
				"managed_by": test.wantManagedBy,
				"tags":       test.wantTags,
			} {
				if got := d.Get(key); !reflect.DeepEqual(got, want) {
					t.Errorf("%s: got %#v, want %#v", key, got, want)
				}
			}

			identity, err := d.Identity()
			if err != nil {
				t.Fatalf("getting identity: %v", err)
			}
			for key, want := range map[string]string{
				"subscription_id": id.SubscriptionId,
				"name":            id.ResourceGroupName,
			} {
				if got := identity.Get(key); got != want {
					t.Errorf("identity %s: got %#v, want %q", key, got, want)
				}
			}
		})
	}
}

func TestResourceGroup2FooSchema(t *testing.T) {
	foo, ok := resourceResourceGroup2().Schema["foo"]
	if !ok {
		t.Fatal("foo is missing from the resource schema")
	}
	if foo.Type != schema.TypeBool || !foo.Optional || !foo.Computed {
		t.Errorf("foo must be an optional, computed boolean: %#v", foo)
	}
}
