package options

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/util"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TerragruntOptions represents options that configure the behavior of the Terragrunt program
type TerragruntOptions struct {
	// Location of the Terragrunt config file
	TerragruntConfigPath string

	// Location of the terraform binary
	TerraformPath string

	// Whether we should prompt the user for confirmation or always assume "yes"
	NonInteractive bool

	// CLI args that are intended for Terraform (i.e. all the CLI args except the --terragrunt ones)
	TerraformCliArgs []string

	// The working directory in which to run Terraform
	WorkingDir string

	// The logger to use for all logging
	Logger *log.Logger

	// Environment variables at runtime
	Env map[string]string

	// Terraform variables at runtime
	Variables VariableList

	// Download Terraform configurations from the specified source location into a temporary folder and run
	// Terraform in that temporary folder
	Source string

	// If set to true, delete the contents of the temporary folder before downloading Terraform source code into it
	SourceUpdate bool

	// If set to true, continue running *-all commands even if a dependency has errors. This is mostly useful for 'output-all <some_variable>'. See https://github.com/gruntwork-io/terragrunt/issues/193
	IgnoreDependencyErrors bool

	// If you want stdout to go somewhere other than os.stdout
	Writer io.Writer

	// If you want stderr to go somewhere other than os.stderr
	ErrWriter io.Writer

	// A command that can be used to run Terragrunt with the given options. This is useful for running Terragrunt
	// multiple times (e.g. when spinning up a stack of Terraform modules). The actual command is normally defined
	// in the cli package, which depends on almost all other packages, so we declare it here so that other
	// packages can use the command without a direct reference back to the cli package (which would create a
	// circular dependency).
	RunTerragrunt func(*TerragruntOptions) error

	// If set in terragrunt configuration, this string is added to the directory name before calculating the hashing
	// This allow differentiation based on certain attribute to ensure that different config (env, region) are executed
	// in distinct folder
	Uniqueness string

	// The list of files (should be only one) where to save files if save_variables() has been invoked by the user
	deferredSaveList map[string]bool

	// Used to replace ${var.xxx} that are not defined by blanks (used in the second parsing pass)
	EraseNonDefinedVariables bool
}

// Create a new TerragruntOptions object with reasonable defaults for real usage
func NewTerragruntOptions(terragruntConfigPath string) *TerragruntOptions {
	workingDir := filepath.Dir(terragruntConfigPath)

	return &TerragruntOptions{
		TerragruntConfigPath:   terragruntConfigPath,
		TerraformPath:          "terraform",
		NonInteractive:         false,
		TerraformCliArgs:       []string{},
		WorkingDir:             workingDir,
		Logger:                 util.CreateLogger(""),
		Env:                    map[string]string{},
		Variables:              VariableList{},
		Source:                 "",
		SourceUpdate:           false,
		IgnoreDependencyErrors: false,
		Writer:                 os.Stdout,
		ErrWriter:              os.Stderr,
		RunTerragrunt: func(terragruntOptions *TerragruntOptions) error {
			return errors.WithStackTrace(RunTerragruntCommandNotSet)
		},
	}
}

// Create a new TerragruntOptions object with reasonable defaults for test usage
func NewTerragruntOptionsForTest(terragruntConfigPath string) *TerragruntOptions {
	opts := NewTerragruntOptions(terragruntConfigPath)

	opts.NonInteractive = true

	return opts
}

// Create a copy of this TerragruntOptions, but with different values for the given variables. This is useful for
// creating a TerragruntOptions that behaves the same way, but is used for a Terraform module in a different folder.
func (terragruntOptions *TerragruntOptions) Clone(terragruntConfigPath string) *TerragruntOptions {
	workingDir := filepath.Dir(terragruntConfigPath)

	newOptions := TerragruntOptions{
		TerragruntConfigPath:   terragruntConfigPath,
		TerraformPath:          terragruntOptions.TerraformPath,
		NonInteractive:         terragruntOptions.NonInteractive,
		TerraformCliArgs:       terragruntOptions.TerraformCliArgs,
		WorkingDir:             workingDir,
		Logger:                 util.CreateLogger(workingDir),
		Env:                    terragruntOptions.Env,
		Variables:              VariableList{},
		Source:                 terragruntOptions.Source,
		SourceUpdate:           terragruntOptions.SourceUpdate,
		IgnoreDependencyErrors: terragruntOptions.IgnoreDependencyErrors,
		Writer:                 terragruntOptions.Writer,
		ErrWriter:              terragruntOptions.ErrWriter,
		RunTerragrunt:          terragruntOptions.RunTerragrunt,
	}

	// We do a deep copy of the variables since they must be disctint from the original
	for key, value := range terragruntOptions.Variables {
		newOptions.Variables.SetValue(key, value.Value, value.Source)
	}
	return &newOptions
}

// SaveVariables - Actually save the variables to the list of deferred files
func (terragruntOptions *TerragruntOptions) SaveVariables() (err error) {
	if terragruntOptions.deferredSaveList != nil {
		variables := make(map[string]interface{}, len(terragruntOptions.Variables))

		// We keep only the value from the variable list, don't need the source
		for key, value := range terragruntOptions.Variables {
			variables[key] = value.Value
		}

		for file := range terragruntOptions.deferredSaveList {
			terragruntOptions.Logger.Printf("Saving variables into %s", file)
			var content []byte
			switch strings.ToLower(filepath.Ext(file)) {
			case ".yml", ".yaml":
				content, err = yaml.Marshal(variables)
				if err != nil {
					return
				}
			default:
				content, err = json.MarshalIndent(variables, "", "  ")
				if err != nil {
					return
				}
			}

			err = ioutil.WriteFile(filepath.Join(terragruntOptions.WorkingDir, file), content, 0644)
		}
	}
	return
}

// AddDeferredSaveVariables - Add a path where to save the variable list
func (terragruntOptions *TerragruntOptions) AddDeferredSaveVariables(filename string) {
	if terragruntOptions.deferredSaveList == nil {
		terragruntOptions.deferredSaveList = map[string]bool{}
	}
	terragruntOptions.deferredSaveList[filename] = true
}

// Custom types
type VariableList map[string]Variable

func (this VariableList) SetValue(key string, value interface{}, source VariableSource) {
	if this[key].Source <= source {
		// We only override value if the source has less or equal precedence than the previous value
		this[key] = Variable{source, value}
	}
}

type VariableSource byte

// Value and origin of a variable (origin is important due to the precedence of the definition)
// i.e. A value specified by -var has precedence over value defined in -var-file
type Variable struct {
	Source VariableSource
	Value  interface{}
}

const (
	UndefinedSource VariableSource = iota
	Default
	Environment
	VarFile
	VarFileExplicit
	VarParameter
	VarParameterExplicit
)

// Custom error types

var RunTerragruntCommandNotSet = fmt.Errorf("The RunTerragrunt option has not been set on this TerragruntOptions object")
