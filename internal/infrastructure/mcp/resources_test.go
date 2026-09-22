package mcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestExtractInstanceID(t *testing.T) {
	cases := map[string]any{
		"uri template match": []string{"1101000001"},
		"string":             "1101000001",
		"number":             float64(1101000001),
	}
	for name, raw := range cases {
		req := mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{
			URI:       "waba://instance/1101000001/state",
			Arguments: map[string]any{"id": raw},
		}}
		id, err := extractInstanceID(req)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if id != 1101000001 {
			t.Errorf("%s: id = %d", name, id)
		}
	}
}

func TestExtractInstanceIDRejectsEmptyMatch(t *testing.T) {
	req := mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{
		URI:       "waba://instance//state",
		Arguments: map[string]any{"id": []string{}},
	}}
	if _, err := extractInstanceID(req); err == nil {
		t.Error("expected an error for an empty {id} match")
	}
}
