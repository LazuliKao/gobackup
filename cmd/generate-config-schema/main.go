package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gobackup/gobackup/config"
	"github.com/invopop/jsonschema"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fatal(err)
	}

	outputPath := filepath.Join(repoRoot, "web", "src", "generated", "config-schema.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		fatal(err)
	}

	reflector := jsonschema.Reflector{
		Anonymous:                 true,
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	schema := reflector.ReflectFromType(reflect.TypeOf(&config.ConfigSchemaSpec{}))
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fatal(err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		fatal(err)
	}
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if isRepoRoot(wd) {
		return wd, nil
	}

	parent := filepath.Dir(wd)
	if isRepoRoot(parent) {
		return parent, nil
	}

	return "", fmt.Errorf("unable to locate repository root from %q", wd)
}

func isRepoRoot(path string) bool {
	if path == "." || path == string(filepath.Separator) {
		return false
	}

	if _, err := os.Stat(filepath.Join(path, "go.mod")); err != nil {
		return false
	}

	if _, err := os.Stat(filepath.Join(path, "config", "schema_spec.go")); err != nil {
		return false
	}

	return true
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
