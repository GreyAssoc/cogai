// cog-gear-lint validates a Tier 1 / Tier 2 cog gear declaration
// against the cog_gear v1 schema. Run before submitting a gear to
// your cog deployment's catalogue; CI integrators run it on PRs.
//
// Usage:
//
//	cog-gear-lint <file.yaml|file.json>
//	cog-gear-lint <dir/>          # validates every *.yaml/*.json in the dir
//
// Exit codes:
//
//	0 — all declarations valid
//	1 — one or more declarations failed validation
//	2 — usage error (missing arg, unreadable file, etc.)
//
// The validator enforces exactly the same envelope rules the cog
// engine's loader applies at runtime. A passing declaration loads
// cleanly; if it doesn't, that's a bug — file an issue at the cogai
// repo.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const Version = "v1"

var nameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)

// allowedMethods, allowedTiers — exact strings the schema accepts.
var allowedMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}
var allowedTiers = map[string]bool{
	"http": true, "webhook": true,
}

// Decl is the minimal parse shape — the validator inspects fields
// directly. Rich type modelling would belong in the engine; this
// CLI stays small.
type Decl struct {
	CogGear          any            `yaml:"cog_gear" json:"cog_gear"`
	Tier             any            `yaml:"tier" json:"tier"`
	Name             any            `yaml:"name" json:"name"`
	Description      any            `yaml:"description" json:"description"`
	Endpoint         map[string]any `yaml:"endpoint" json:"endpoint"`
	InputSchema      any            `yaml:"input_schema" json:"input_schema"`
	OutputSchema     any            `yaml:"output_schema,omitempty" json:"output_schema,omitempty"`
	Permissions      map[string]any `yaml:"permissions" json:"permissions"`
	ExternalWorkflow map[string]any `yaml:"external_workflow,omitempty" json:"external_workflow,omitempty"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Printf("cog-gear-lint %s\n", Version)
		os.Exit(0)
	}
	if args[0] == "--help" || args[0] == "-h" {
		usage()
		os.Exit(0)
	}

	failed := 0
	for _, arg := range args {
		paths, err := expand(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", arg, err)
			failed++
			continue
		}
		for _, p := range paths {
			if err := validateFile(p); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", p, err)
				failed++
				continue
			}
			fmt.Printf("OK: %s\n", p)
		}
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: cog-gear-lint <file.yaml|file.json|dir>...")
	fmt.Fprintln(os.Stderr, "Validates Tier 1 / Tier 2 cog gear declarations against cog_gear v1.")
}

// expand turns a CLI argument (file or directory) into a list of
// files to validate. Directories scan one level for .yaml/.yml/.json.
func expand(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" {
			out = append(out, filepath.Join(path, e.Name()))
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("directory contains no .yaml/.yml/.json files")
	}
	return out, nil
}

func validateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(path))
	var d Decl
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &d); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &d); err != nil {
			return fmt.Errorf("parse YAML: %w", err)
		}
	default:
		return fmt.Errorf("unknown extension %q (expected .yaml, .yml, or .json)", ext)
	}
	return validate(d)
}

// validate checks the declaration against the cog_gear v1 schema.
// Returns the first error found — fix one at a time. Multi-error
// reporting can land if the user volume warrants it.
func validate(d Decl) error {
	if err := mustStringEq(d.CogGear, "cog_gear", Version); err != nil {
		return err
	}
	tier, err := mustString(d.Tier, "tier")
	if err != nil {
		return err
	}
	if !allowedTiers[tier] {
		return fmt.Errorf("tier: %q is not allowed (want http or webhook)", tier)
	}
	name, err := mustString(d.Name, "name")
	if err != nil {
		return err
	}
	if !nameRE.MatchString(name) {
		return fmt.Errorf("name: %q does not match ^[a-z][a-z0-9_]{2,63}$ (lowercase snake_case, 3-64 chars)", name)
	}
	desc, err := mustString(d.Description, "description")
	if err != nil {
		return err
	}
	if len(desc) < 10 {
		return fmt.Errorf("description: too short (want >= 10 chars)")
	}
	if len(desc) > 1024 {
		return fmt.Errorf("description: too long (want <= 1024 chars)")
	}

	if d.Endpoint == nil {
		return errors.New("endpoint: required object missing")
	}
	method, ok := d.Endpoint["method"].(string)
	if !ok || method == "" {
		return errors.New("endpoint.method: required string missing")
	}
	if !allowedMethods[method] {
		return fmt.Errorf("endpoint.method: %q not in GET/POST/PUT/PATCH/DELETE", method)
	}
	urlStr, ok := d.Endpoint["url"].(string)
	if !ok || urlStr == "" {
		return errors.New("endpoint.url: required string missing")
	}
	urlHost, err := hostFromTemplatedURL(urlStr)
	if err != nil {
		return fmt.Errorf("endpoint.url: %w", err)
	}

	if d.InputSchema == nil {
		return errors.New("input_schema: required object missing")
	}

	if d.Permissions == nil {
		return errors.New("permissions: required object missing")
	}
	if err := validatePermissions(d.Permissions, tier, urlHost); err != nil {
		return err
	}

	switch tier {
	case "http":
		if d.ExternalWorkflow != nil {
			return errors.New("external_workflow: forbidden for tier=http")
		}
	case "webhook":
		if d.ExternalWorkflow == nil {
			return errors.New("external_workflow: required for tier=webhook")
		}
		if _, ok := d.ExternalWorkflow["platform"].(string); !ok {
			return errors.New("external_workflow.platform: required string missing")
		}
	}

	return nil
}

// validatePermissions enforces the Tier 1 / Tier 2 envelope rules:
// network[] non-empty; timeout_seconds within bounds; no subprocess,
// file_read, file_write.
func validatePermissions(perms map[string]any, tier, urlHost string) error {
	netRaw, ok := perms["network"]
	if !ok {
		return errors.New("permissions.network: required, must list host/port pairs")
	}
	netList, ok := netRaw.([]any)
	if !ok || len(netList) == 0 {
		return errors.New("permissions.network: must be a non-empty array of {host, port}")
	}
	hostFound := false
	for i, item := range netList {
		m, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("permissions.network[%d]: not an object", i)
		}
		host, ok := m["host"].(string)
		if !ok || host == "" {
			return fmt.Errorf("permissions.network[%d].host: required string missing", i)
		}
		if host == urlHost {
			hostFound = true
		}
		port, ok := asInt(m["port"])
		if !ok || port < 1 || port > 65535 {
			return fmt.Errorf("permissions.network[%d].port: must be integer 1..65535", i)
		}
	}
	if urlHost != "" && !hostFound {
		return fmt.Errorf("permissions.network: URL host %q not declared", urlHost)
	}

	timeoutRaw, ok := perms["timeout_seconds"]
	if !ok {
		return errors.New("permissions.timeout_seconds: required integer missing")
	}
	timeout, ok := asInt(timeoutRaw)
	if !ok || timeout < 1 || timeout > 60 {
		return errors.New("permissions.timeout_seconds: must be integer 1..60 (engine default cap)")
	}

	if sub, ok := perms["subprocess"]; ok {
		if b, _ := sub.(bool); b {
			return fmt.Errorf("permissions.subprocess: forbidden in tier=%s", tier)
		}
	}
	if fr, ok := perms["file_read"]; ok {
		if arr, _ := fr.([]any); len(arr) > 0 {
			return fmt.Errorf("permissions.file_read: forbidden in tier=%s", tier)
		}
	}
	if fw, ok := perms["file_write"]; ok {
		if arr, _ := fw.([]any); len(arr) > 0 {
			return fmt.Errorf("permissions.file_write: forbidden in tier=%s", tier)
		}
	}

	return nil
}

// hostFromTemplatedURL extracts the host portion of an endpoint URL.
// URLs may contain {{ }} templates in the path / query; the host
// portion is required to be literal (the cog engine treats it as
// immutable at runtime).
func hostFromTemplatedURL(s string) (string, error) {
	if strings.Contains(beforePath(s), "{{") {
		return "", errors.New("host portion must not contain templates")
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return "", fmt.Errorf("scheme must be https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("URL must have a host")
	}
	host := u.Hostname()
	return host, nil
}

// beforePath returns the prefix of the URL up to (but not including)
// the path. Used to check the host portion for templates without
// pulling in a full URL parser for malformed inputs.
func beforePath(s string) string {
	i := strings.Index(s, "://")
	if i < 0 {
		return s
	}
	rest := s[i+3:]
	j := strings.IndexAny(rest, "/?#")
	if j < 0 {
		return s
	}
	return s[:i+3+j]
}

// asInt accepts int, float64, json.Number — all the shapes YAML
// and JSON unmarshalling may produce for integer-typed fields.
func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		if x == float64(int(x)) {
			return int(x), true
		}
		return 0, false
	case json.Number:
		i, err := x.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	}
	return 0, false
}

func mustString(v any, field string) (string, error) {
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("%s: required string missing", field)
	}
	return s, nil
}

func mustStringEq(v any, field, want string) error {
	s, err := mustString(v, field)
	if err != nil {
		return err
	}
	if s != want {
		return fmt.Errorf("%s: %q != %q", field, s, want)
	}
	return nil
}
