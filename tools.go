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

// ParseToolSchemas turns an OpenAPI 3 document into tool schemas (one per
// operationId). PathPrefix, when non-empty, keeps only that path subtree.
// Operations with explicit empty security (public) are skipped, matching
// NewHandler.
func ParseToolSchemas(openAPIYAML []byte, pathPrefix string) ([]ToolSchema, error) {
	if len(openAPIYAML) == 0 {
		return nil, fmt.Errorf("mcpopenapi: OpenAPIYAML is required")
	}
	ops, err := parseOperations(openAPIYAML, pathPrefix)
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
