package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_MinimalValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	if err := os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: echo_status
description: Echo the HTTP status code from httpbin.
endpoint:
  method: GET
  url: "https://httpbin.org/status/200"
input_schema:
  type: object
permissions:
  network:
    - host: httpbin.org
      port: 443
  timeout_seconds: 5
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := validateFile(path); err != nil {
		t.Errorf("expected valid, got %v", err)
	}
}

func TestValidate_RejectsBadName(t *testing.T) {
	cases := []string{
		"X",                      // uppercase
		"ab",                     // too short
		strings.Repeat("a", 100), // too long
		"123_starts_with_digit",
		"has-hyphen",
		"has space",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "g.yaml")
			content := `
cog_gear: v1
tier: http
name: "` + name + `"
description: Some valid description over ten chars.
endpoint:
  method: GET
  url: "https://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: api.example.com
      port: 443
  timeout_seconds: 5
`
			_ = os.WriteFile(path, []byte(content), 0644)
			err := validateFile(path)
			if err == nil {
				t.Errorf("expected error for name %q, got nil", name)
			} else if !strings.Contains(err.Error(), "name") {
				t.Errorf("expected name error, got %v", err)
			}
		})
	}
}

func TestValidate_RejectsSubprocess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: ok_name
description: Description that is over ten chars long.
endpoint:
  method: GET
  url: "https://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: api.example.com
      port: 443
  timeout_seconds: 5
  subprocess: true
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "subprocess") {
		t.Errorf("expected subprocess rejection, got %v", err)
	}
}

func TestValidate_RejectsFileRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: ok_name
description: Description that is over ten chars long.
endpoint:
  method: GET
  url: "https://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: api.example.com
      port: 443
  timeout_seconds: 5
  file_read: ["/etc"]
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "file_read") {
		t.Errorf("expected file_read rejection, got %v", err)
	}
}

func TestValidate_RejectsHostMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: ok_name
description: Description that is over ten chars long.
endpoint:
  method: GET
  url: "https://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: other.example.com
      port: 443
  timeout_seconds: 5
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "URL host") {
		t.Errorf("expected host mismatch error, got %v", err)
	}
}

func TestValidate_RejectsHTTP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: ok_name
description: Description that is over ten chars long.
endpoint:
  method: GET
  url: "http://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: api.example.com
      port: 80
  timeout_seconds: 5
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "https") {
		t.Errorf("expected https requirement, got %v", err)
	}
}

func TestValidate_WebhookRequiresExternalWorkflow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: webhook
name: webhook_gear
description: Webhook gear without external_workflow block.
endpoint:
  method: POST
  url: "https://hooks.zapier.com/hooks/catch/abc"
input_schema:
  type: object
permissions:
  network:
    - host: hooks.zapier.com
      port: 443
  timeout_seconds: 5
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "external_workflow") {
		t.Errorf("expected external_workflow requirement, got %v", err)
	}
}

func TestValidate_HTTPForbidsExternalWorkflow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: ok_name
description: Description over ten chars long.
endpoint:
  method: GET
  url: "https://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: api.example.com
      port: 443
  timeout_seconds: 5
external_workflow:
  platform: zapier
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "external_workflow") {
		t.Errorf("expected forbidden external_workflow, got %v", err)
	}
}

func TestValidate_TimeoutOverCap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.yaml")
	_ = os.WriteFile(path, []byte(`
cog_gear: v1
tier: http
name: ok_name
description: Description over ten chars long.
endpoint:
  method: GET
  url: "https://api.example.com/v1/x"
input_schema:
  type: object
permissions:
  network:
    - host: api.example.com
      port: 443
  timeout_seconds: 120
`), 0644)
	err := validateFile(path)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout cap error, got %v", err)
	}
}

func TestValidate_Examples(t *testing.T) {
	// The four shipped example gears should all validate clean.
	// Run from the cog-gear-lint/ directory so ../examples resolves.
	wd, _ := os.Getwd()
	examples := filepath.Join(wd, "..", "examples")
	if _, err := os.Stat(examples); err != nil {
		t.Skip("examples/ not adjacent; running outside the cogai repo layout")
	}
	entries, err := os.ReadDir(examples)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			if err := validateFile(filepath.Join(examples, e.Name())); err != nil {
				t.Errorf("%s did not validate: %v", e.Name(), err)
			}
		})
	}
}
