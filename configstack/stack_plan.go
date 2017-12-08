package configstack

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/shell"
	"github.com/gruntwork-io/terragrunt/util"
)

// The returned information for each module
type moduleResult struct {
	Module    TerraformModule
	Err       error
	Message   string
	NbChanges int
}

var planResultRegex = regexp.MustCompile(`(\d+) to add, (\d+) to change, (\d+) to destroy.`)

func (stack *Stack) planWithSummary(terragruntOptions *options.TerragruntOptions) error {
	// We override the multi errors creator to use a specialized error type for plan
	// because error severity in plan is not standard (i.e. exit code 2 is less significant that exit code 1).
	CreateMultiErrors = func(errs []error) error {
		return PlanMultiError{MultiError{errs}}
	}

	// We do a special treatment for -detailed-exitcode since we do not want to interrupt the processing of dependant
	// stacks if one dependency has changes
	detailedExitCode := util.ListContainsElement(terragruntOptions.TerraformCliArgs, "-detailed-exitcode")
	if detailedExitCode {
		util.RemoveElementFromList(terragruntOptions.TerraformCliArgs, "-detailed-exitcode")
	}

	results := make([]moduleResult, 0, len(stack.Modules))
	err := RunModulesWithHandler(stack.Modules, getResultHandler(detailedExitCode, &results), NormalOrder)
	printSummary(terragruntOptions, results)

	// If there is no error, but -detail-exitcode is specified, we return an error with the number of changes.
	if err == nil && detailedExitCode {
		sum := 0
		for _, status := range results {
			sum += status.NbChanges
		}
		if sum != 0 {
			article, plural := "is", ""
			if sum > 1 {
				article, plural = "are", "s"
			}
			terragruntOptions.Logger.Noticef("There %s %v change%s to apply", article, sum, plural)
			return errors.PlanWithChanges{}
		}
	}

	return err
}

// Returns the handler that will be executed after each completion of `terraform plan`
func getResultHandler(detailedExitCode bool, results *[]moduleResult) ModuleHandler {
	return func(module TerraformModule, output string, err error) (string, error) {
		warnAboutMissingDependencies(module, output)
		if exitCode, convErr := shell.GetExitCode(err); convErr == nil && detailedExitCode && exitCode == errors.CHANGE_EXIT_CODE {
			// We do not want to consider CHANGE_EXIT_CODE as an error and not execute the dependants because there is an "error" in the dependencies.
			// CHANGE_EXIT_CODE is not an error in this case, it is simply a status. We will reintroduce the exit code at the very end to mimic the behaviour
			// of the native terrafrom plan -detailed-exitcode to exit with CHANGE_EXIT_CODE if there are changes in any of the module in the stack.
			err = nil
		}

		if output != "" {
			message, count := extractSummaryResultFromPlan(output)

			// We add the result to the result list (there is no concurrency problem because it is handled by the running_module)
			*results = append(*results, moduleResult{module, err, message, count})
		}

		return output, err
	}
}

// Print a little summary of the plan execution
func printSummary(terragruntOptions *options.TerragruntOptions, results []moduleResult) {
	terragruntOptions.Printf("%s\nSummary:\n", separator)

	var length int
	for _, result := range results {
		nameLength := len(util.GetPathRelativeToWorkingDir(result.Module.Path))
		if nameLength > length {
			length = nameLength
		}
	}

	format := fmt.Sprintf("    %%-%dv : %%v%%v\n", length)
	for _, result := range results {
		errMsg := ""
		if result.Err != nil {
			errMsg = fmt.Sprintf(", Error: %v", result.Err)
		}

		terragruntOptions.Printf(format, util.GetPathRelativeToWorkingDir(result.Module.Path), result.Message, errMsg)
	}
}

// Check the output message
func warnAboutMissingDependencies(module TerraformModule, output string) {
	if strings.Contains(output, ": Resource 'data.terraform_remote_state.") {
		var dependenciesMsg string
		if len(module.Dependencies) > 0 {
			dependenciesMsg = fmt.Sprintf(" contains dependencies to %v and", module.Config.Dependencies.Paths)
		}
		module.TerragruntOptions.Logger.Warningf("%v%v refers to remote state, you may have to apply your changes in the dependencies prior running terragrunt plan-all.\n",
			module.Path,
			dependenciesMsg,
		)
	}
}

// Parse the output message to extract a summary
func extractSummaryResultFromPlan(output string) (string, int) {
	const noChange = "No changes. Infrastructure is up-to-date."
	if strings.Contains(output, noChange) {
		return "No change", 0
	}

	result := planResultRegex.FindStringSubmatch(output)
	if len(result) == 0 {
		return "Unable to determine the plan status", -1
	}

	// Count the total number of changes
	sum := 0
	for _, value := range result[1:] {
		count, _ := strconv.Atoi(value)
		sum += count
	}
	if sum != 0 {
		return result[0], sum
	}

	// Sometimes, terraform returns 0 add, 0 change and 0 destroy. We return a more explicit message
	return "No effective change", 0
}

// This is a specialized version of MultiError type
// It handles the exit code differently from the base implementation
type PlanMultiError struct {
	MultiError
}

func (this PlanMultiError) ExitStatus() (int, error) {
	exitCode := NORMAL_EXIT_CODE
	for i := range this.Errors {
		if code, err := shell.GetExitCode(this.Errors[i]); err != nil {
			return UNDEFINED_EXIT_CODE, this
		} else if code == ERROR_EXIT_CODE || code == errors.CHANGE_EXIT_CODE && exitCode == NORMAL_EXIT_CODE {
			// The exit code 1 is more significant that the exit code 2 because it represents an error
			// while 2 represent a warning.
			return UNDEFINED_EXIT_CODE, this
		}
	}
	return exitCode, nil
}
