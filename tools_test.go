package mcpopenapi_test

import (
	"testing"

	"github.com/yaronf/mcpopenapi"
)

func TestParseToolSchemas(t *testing.T) {
	tools, err := mcpopenapi.ParseToolSchemas([]byte(sampleOpenAPI), "/api/agent")
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]mcpopenapi.ToolSchema{}
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	if _, ok := byName["health"]; ok {
		t.Fatal("public health should be skipped")
	}
	getTrip, ok := byName["getTrip"]
	if !ok {
		t.Fatalf("missing getTrip; got %#v", byName)
	}
	if !getTrip.ReadOnly {
		t.Fatal("getTrip should be read-only")
	}
	props, _ := getTrip.InputSchema["properties"].(map[string]any)
	if _, ok := props["id"]; !ok {
		t.Fatalf("getTrip input missing id: %#v", getTrip.InputSchema)
	}
	patch, ok := byName["patchTrip"]
	if !ok {
		t.Fatal("missing patchTrip")
	}
	if patch.ReadOnly {
		t.Fatal("patchTrip should not be read-only")
	}
}
