package infrastructure

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postRegister(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	o := NewOAuthServer("https://mcp.example.com", nil, nil)
	rec := httptest.NewRecorder()
	o.RegisterHandler()(rec, httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body)))
	return rec
}

func TestRegisterIssuesPublicClient(t *testing.T) {
	rec := postRegister(t, `{"client_name":"Claude","redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"token_endpoint_auth_method":"none"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ClientID                string   `json:"client_id"`
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		ClientSecret            string   `json:"client_secret"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ClientID == "" {
		t.Error("client_id is empty")
	}
	if resp.ClientSecret != "" {
		t.Error("a public client must not get a client_secret")
	}
	if resp.TokenEndpointAuthMethod != "none" {
		t.Errorf("token_endpoint_auth_method = %q", resp.TokenEndpointAuthMethod)
	}
	if resp.ClientName != "Claude" || len(resp.RedirectURIs) != 1 || resp.RedirectURIs[0] != "https://claude.ai/api/mcp/auth_callback" {
		t.Errorf("metadata not echoed: %+v", resp)
	}
}

func TestRegisterIssuesDistinctClientIDs(t *testing.T) {
	body := `{"redirect_uris":["http://127.0.0.1:6274/oauth/callback"]}`
	first, second := postRegister(t, body), postRegister(t, body)
	if first.Body.String() == second.Body.String() {
		t.Error("two registrations returned the same client")
	}
}

func TestRegisterRejectsInvalidMetadata(t *testing.T) {
	cases := map[string]string{
		"not json":            `redirect_uris=x`,
		"no redirect uris":    `{"client_name":"x"}`,
		"relative redirect":   `{"redirect_uris":["/callback"]}`,
		"confidential client": `{"redirect_uris":["https://claude.ai/cb"],"token_endpoint_auth_method":"client_secret_basic"}`,
	}
	for name, body := range cases {
		rec := postRegister(t, body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", name, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"error":"invalid_`) {
			t.Errorf("%s: body = %s", name, rec.Body.String())
		}
	}
}

func TestRegisterRejectsGet(t *testing.T) {
	o := NewOAuthServer("https://mcp.example.com", nil, nil)
	rec := httptest.NewRecorder()
	o.RegisterHandler()(rec, httptest.NewRequest(http.MethodGet, "/register", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestMetadataAdvertisesRegistrationEndpoint(t *testing.T) {
	o := NewOAuthServer("https://mcp.example.com", nil, nil)
	rec := httptest.NewRecorder()
	o.MetadataHandler()(rec, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil))
	var meta map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta["registration_endpoint"] != "https://mcp.example.com/register" {
		t.Errorf("registration_endpoint = %v", meta["registration_endpoint"])
	}
}
