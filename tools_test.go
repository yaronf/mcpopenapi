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

const audienceOpenAPI = `
openapi: 3.1.0
info:
  title: test
  version: 0.0.1
paths:
  /api/shared:
    get:
      operationId: sharedOp
      summary: Shared
      x-audiences: [api, assistant]
      security:
        - bearerAuth: []
      responses:
        "200":
          description: OK
  /api/assistant-only:
    post:
      operationId: assistantOp
      summary: Assistant only
      x-audiences: [assistant]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                q:
                  type: string
      responses:
        "200":
          description: OK
  /api/unannotated:
    get:
      operationId: plainOp
      summary: No audiences
      security:
        - bearerAuth: []
      responses:
        "200":
          description: OK
`

func TestParseToolSchemasAudience(t *testing.T) {
	assistant, err := mcpopenapi.ParseToolSchemasOpts([]byte(audienceOpenAPI), mcpopenapi.ParseOptions{
		Audience:           "assistant",
		IncludeUnannotated: mcpopenapi.Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range assistant {
		names[tool.Name] = true
	}
	if !names["sharedOp"] || !names["assistantOp"] {
		t.Fatalf("assistant filter missing ops: %#v", names)
	}
	if names["plainOp"] {
		t.Fatal("unannotated should be excluded when IncludeUnannotated=false")
	}

	apiSide, err := mcpopenapi.ParseToolSchemasOpts([]byte(audienceOpenAPI), mcpopenapi.ParseOptions{
		Audience: "api",
		// IncludeUnannotated default true
	})
	if err != nil {
		t.Fatal(err)
	}
	names = map[string]bool{}
	for _, tool := range apiSide {
		names[tool.Name] = true
	}
	if !names["sharedOp"] || !names["plainOp"] {
		t.Fatalf("api filter missing ops: %#v", names)
	}
	if names["assistantOp"] {
		t.Fatal("assistant-only op should be excluded for audience=api")
	}
}
