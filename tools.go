package mcpopenapi

import "fmt"

// ToolSchema is an OpenAPI-derived tool definition usable by MCP or other
// tool runtimes (e.g. OpenAI function calling).
type ToolSchema struct {
	Name        string
	Description string
	InputSchema map[string]any
	ReadOnly    bool
}

// ParseOptions configures OpenAPI→tool schema extraction.
type ParseOptions struct {
	// PathPrefix, when non-empty, keeps only that path subtree.
	PathPrefix string
	// Audience, when non-empty, filters by the audience extension on each
	// operation (see AudienceKey).
	Audience string
	// AudienceKey is the OpenAPI extension key listing audiences for an
	// operation. Empty means "x-audiences".
	AudienceKey string
	// IncludeUnannotated controls operations that omit the audience
	// extension when Audience is set. nil means true (include them).
	// Set to false to keep only operations that explicitly list Audience.
	IncludeUnannotated *bool
}

// ParseToolSchemas turns an OpenAPI 3 document into tool schemas (one per
// operationId). PathPrefix, when non-empty, keeps only that path subtree.
// Operations with explicit empty security (public) are skipped, matching
// NewHandler.
func ParseToolSchemas(openAPIYAML []byte, pathPrefix string) ([]ToolSchema, error) {
	return ParseToolSchemasOpts(openAPIYAML, ParseOptions{PathPrefix: pathPrefix})
}

// ParseToolSchemasOpts is like ParseToolSchemas with audience filtering.
func ParseToolSchemasOpts(openAPIYAML []byte, opts ParseOptions) ([]ToolSchema, error) {
	if len(openAPIYAML) == 0 {
		return nil, fmt.Errorf("mcpopenapi: OpenAPIYAML is required")
	}
	ops, err := parseOperations(openAPIYAML, opts)
	if err != nil {
		return nil, err
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("mcpopenapi: no operations found")
	}
	out := make([]ToolSchema, 0, len(ops))
	for _, op := range ops {
		out = append(out, ToolSchema{
			Name:        op.ID,
			Description: toolDescription(op.Summary, op.Description),
			InputSchema: op.InputSchema,
			ReadOnly:    op.ReadOnly,
		})
	}
	return out, nil
}

// Bool returns a pointer to b for ParseOptions / Config fields.
func Bool(b bool) *bool { return &b }
