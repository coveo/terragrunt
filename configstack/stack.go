package configstack

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
)

// Stack represents a stack of Terraform modules (i.e. folders with Terraform templates) that you can "spin up" or
// "spin down" in a single command
type Stack struct {
	Path    string
	Modules []*TerraformModule
}

// Render this stack as a human-readable string
func (stack *Stack) String() string {
	modules := []string{}
	for _, module := range stack.Modules {
		modules = append(modules, fmt.Sprintf("  => %s", module.String()))
	}
	return fmt.Sprintf("Stack at %s:\n%s", stack.Path, strings.Join(modules, "\n"))
}

// SimpleModules returns the list of modules (simplified serializable version)
func (stack Stack) SimpleModules() SimpleTerraformModules {
	modules := make(SimpleTerraformModules, len(stack.Modules))
	for i := range stack.Modules {
		modules[i] = stack.Modules[i].Simple()
	}
	return modules
}

// JSON renders this stack as a JSON string
func (stack Stack) JSON() string {
	json, err := json.MarshalIndent(stack.SimpleModules(), "", "  ")
	if err != nil {
		panic(err)
	}
	return string(json)
}

// SortModules sorts in-place the list of modules topologically
func (stack *Stack) SortModules() {
	sort.Slice(stack.Modules, func(i, j int) bool {
		return stack.Modules[i].Path < stack.Modules[j].Path
	})
	sortedModules := make([]*TerraformModule, 0)
	visitedModules := make(map[string]bool)
	for _, module := range stack.Modules {
		if _, ok := visitedModules[module.Path]; !ok {
			visitedModules, sortedModules = stack.topologicalSort(module, visitedModules, sortedModules)
		}
	}
	stack.Modules = sortedModules
}

func (stack *Stack) topologicalSort(module *TerraformModule, visitedModules map[string]bool, sortedModules []*TerraformModule) (map[string]bool, []*TerraformModule) {
	visitedModules[module.Path] = true
	for _, dependency := range module.Dependencies {
		if _, ok := visitedModules[dependency.Path]; !ok {
			visitedModules, sortedModules = stack.topologicalSort(dependency, visitedModules, sortedModules)
		}
	}
	return visitedModules, append(sortedModules, module)
}

// Plan all the modules in the given stack in their specified order.
func (stack *Stack) Plan(command string, terragruntOptions *options.TerragruntOptions) error {
	stack.setTerraformCommand([]string{command})
	return stack.planWithSummary(terragruntOptions)
}

// Output prints the outputs of all the modules in the given stack in their specified order.
func (stack *Stack) Output(command string, terragruntOptions *options.TerragruntOptions) error {
	stack.setTerraformCommand([]string{command})
	handler := func(module TerraformModule, output string, err error) (string, error) {
		if err != nil && strings.Contains(output, "no outputs defined") {
			return "", nil
		}
		return output, err
	}
	return runModulesWithHandler(stack.Modules, handler, NormalOrder)
}

// RunAll runs the specified command on all modules in the given stack in their specified order.
func (stack *Stack) RunAll(command []string, terragruntOptions *options.TerragruntOptions, order dependencyOrder) error {
	stack.setTerraformCommand(command)
	return runModulesWithHandler(stack.Modules, nil, order)
}

// Return an error if there is a dependency cycle in the modules of this stack.
func (stack *Stack) checkForCycles() error {
	return checkForCycles(stack.Modules)
}

// FindStackInSubfolders finds all the Terraform modules in the subfolders of the working directory of the given TerragruntOptions and
// assemble them into a Stack object that can be applied or destroyed in a single command
func FindStackInSubfolders(terragruntOptions *options.TerragruntOptions) (*Stack, error) {
	terragruntConfigFiles, err := config.FindConfigFilesInPath(terragruntOptions)
	if err != nil {
		return nil, err
	}

	return createStackForTerragruntConfigPaths(terragruntOptions.WorkingDir, terragruntConfigFiles, terragruntOptions)
}

// Set the command in the TerragruntOptions object of each module in this stack to the given command.
func (stack *Stack) setTerraformCommand(command []string) {
	for _, module := range stack.Modules {
		// We duplicate the args to make sure that each module gets its own copy of the args
		newArgs := make([]string, len(command))
		copy(newArgs, command)
		module.TerragruntOptions.TerraformCliArgs = append(newArgs, module.TerragruntOptions.TerraformCliArgs...)
	}
}

// Find all the Terraform modules in the folders that contain the given Terragrunt config files and assemble those
// modules into a Stack object that can be applied or destroyed in a single command
func createStackForTerragruntConfigPaths(path string, terragruntConfigPaths []string, terragruntOptions *options.TerragruntOptions) (*Stack, error) {
	if len(terragruntConfigPaths) == 0 {
		terragruntOptions.Logger.Warning("Could not find any subfolders with Terragrunt configuration files")
	}

	modules, err := ResolveTerraformModules(terragruntConfigPaths, terragruntOptions)
	if err != nil {
		return nil, err
	}

	stack := &Stack{Path: path, Modules: modules}
	if err := stack.checkForCycles(); err != nil {
		return nil, err
	}

	return stack, nil
}

// Custom error types

type errDependencyCycle []string

func (err errDependencyCycle) Error() string {
	return fmt.Sprintf("Found a dependency cycle between modules: %s", strings.Join([]string(err), " -> "))
}
