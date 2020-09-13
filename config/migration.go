package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

// MigrationFile represents a config for migration written in HCL.
type MigrationFile struct {
	// Migration is a migration block.
	// It must contain only one block, and multiple blocks are not allowed,
	// because it's hard to re-run the file if partially failed.
	Migration MigrationBlock `hcl:"migration,block"`
}

// MigrationBlock represents a migration block in HCL.
type MigrationBlock struct {
	// Type is a type for migration.
	// Valid values are `state` and `multi_state`.
	Type string `hcl:"type,label"`
	// Name is an arbitrary name for migration.
	Name string `hcl:"name,label"`
	// Remain is a body of migration block.
	// We first decode only a block header and then decode schema depending on
	// its type label.
	Remain hcl.Body `hcl:",remain"`
}

// ParseMigrationFile parses a given source of migration file and returns a *tfmigrate.MigrationConfig.
// Note that this method does not read a file and you should pass source of config in bytes.
// The filename is used for error message and selecting HCL syntax (.hcl and .json).
func ParseMigrationFile(filename string, source []byte) (*tfmigrate.MigrationConfig, error) {
	// Decode migration block header.
	var f MigrationFile
	err := hclsimple.Decode(filename, source, nil, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode migration file: %s, err: %s", filename, err)
	}

	migrator, err := parseMigrationBlock(f.Migration)
	if err != nil {
		return nil, err
	}

	config := &tfmigrate.MigrationConfig{
		Type:     f.Migration.Type,
		Name:     f.Migration.Name,
		Migrator: migrator,
	}

	return config, nil
}

// parseMigrationBlock parses a migration block and returns a tfmigrate.MigratorConfig.
func parseMigrationBlock(b MigrationBlock) (tfmigrate.MigratorConfig, error) {
	switch b.Type {
	case "state":
		return parseStateMigrationBlock(b)

	case "multi_state":
		return parseMultiStateMigrationBlock(b)

	default:
		return nil, fmt.Errorf("unknown migration type: %s", b.Type)
	}
}

// parseStateMigrationBlock parses a migration block for state and returns a tfmigrate.MigratorConfig.
func parseStateMigrationBlock(b MigrationBlock) (tfmigrate.MigratorConfig, error) {
	var config tfmigrate.StateMigratorConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}

// parseMultiStateMigrationBlock parses a migration block for multi_state and
// returns a tfmigrate.MigratorConfig.
func parseMultiStateMigrationBlock(b MigrationBlock) (tfmigrate.MigratorConfig, error) {
	var config tfmigrate.MultiStateMigratorConfig
	diags := gohcl.DecodeBody(b.Remain, nil, &config)
	if diags.HasErrors() {
		return nil, diags
	}

	return &config, nil
}
