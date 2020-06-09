package config

import (
	"testing"

	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/test/helpers"
	"github.com/stretchr/testify/assert"
)

var mockDefaultInclude = IncludeConfig{Path: DefaultTerragruntConfigPath}

func TestPathRelativeToInclude(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		include           IncludeConfig
		terragruntOptions *options.TerragruntOptions
		expectedPath      string
	}{
		{
			mockDefaultInclude,
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child"),
			".",
		},
		{
			IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/.terragrunt"),
			"child",
		},
		{
			IncludeConfig{Path: helpers.RootFolder + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child"),
			"child",
		},
		{
			IncludeConfig{Path: "../../../" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/sub-child/sub-sub-child"),
			"child/sub-child/sub-sub-child",
		},
		{
			IncludeConfig{Path: helpers.RootFolder + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/sub-child/sub-sub-child"),
			"child/sub-child/sub-sub-child",
		},
		{
			IncludeConfig{Path: "../../other-child/" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/sub-child"),
			"../child/sub-child",
		},
		{
			IncludeConfig{Path: "../../" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest("../child/sub-child"),
			"child/sub-child",
		},
		{
			IncludeConfig{Path: "find_in_parent_folders()"},
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child"),
			"child/sub-child",
		},
	}

	for _, testCase := range testCases {
		context := resolveContext{include: testCase.include, options: testCase.terragruntOptions}
		if context.include.Path == "find_in_parent_folders()" {
			path, _ := context.findInParentFolders()
			context.include.Path = path
		}
		actualPath, actualErr := context.pathRelativeToInclude()
		assert.Nil(t, actualErr, "For include %v and options %v, unexpected error: %v", testCase.include, testCase.terragruntOptions, actualErr)
		assert.Equal(t, testCase.expectedPath, actualPath, "For include %v and options %v", testCase.include, testCase.terragruntOptions)
	}
}

func TestPathRelativeFromInclude(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		include           IncludeConfig
		terragruntOptions *options.TerragruntOptions
		expectedPath      string
	}{
		{
			mockDefaultInclude,
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child"),
			".",
		},
		{
			IncludeConfig{Path: "../" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child"),
			"..",
		},
		{
			IncludeConfig{Path: helpers.RootFolder + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child"),
			"..",
		},
		{
			IncludeConfig{Path: "../../../" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/sub-child/sub-sub-child"),
			"../../..",
		},
		{
			IncludeConfig{Path: helpers.RootFolder + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/sub-child/sub-sub-child"),
			"../../..",
		},
		{
			IncludeConfig{Path: "../../other-child/" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest(helpers.RootFolder + "child/sub-child"),
			"../../other-child",
		},
		{
			IncludeConfig{Path: "../../" + DefaultTerragruntConfigPath},
			options.NewTerragruntOptionsForTest("../child/sub-child"),
			"../..",
		},
		{
			IncludeConfig{Path: "find_in_parent_folders()"},
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child"),
			"../..",
		},
	}

	for _, testCase := range testCases {
		context := resolveContext{include: testCase.include, options: testCase.terragruntOptions}
		if context.include.Path == "find_in_parent_folders()" {
			path, _ := context.findInParentFolders()
			context.include.Path = path
		}
		actualPath, actualErr := context.pathRelativeFromInclude()
		assert.Nil(t, actualErr, "For include %v and options %v, unexpected error: %v", testCase.include, testCase.terragruntOptions, actualErr)
		assert.Equal(t, testCase.expectedPath, actualPath, "For include %v and options %v", testCase.include, testCase.terragruntOptions)
	}
}

func TestFindInParentFolders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		terragruntOptions *options.TerragruntOptions
		expectedPath      string
		expectedErr       error
	}{
		{
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child"),
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/terragrunt-in-root/child/sub-child/sub-sub-child"),
			"../../../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child"),
			"",
			parentTerragruntConfigNotFound("../test/fixture-parent-folders/no-terragrunt-in-root/child/sub-child/" + DefaultTerragruntConfigPath),
		},
		{
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/multiple-terragrunt-in-parents/child"),
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/sub-child"),
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.NewTerragruntOptionsForTest("../test/fixture-parent-folders/multiple-terragrunt-in-parents/child/sub-child/sub-sub-child" + DefaultTerragruntConfigPath),
			"../" + DefaultTerragruntConfigPath,
			nil,
		},
		{
			options.NewTerragruntOptionsForTest("/"),
			"",
			parentTerragruntConfigNotFound("/" + DefaultTerragruntConfigPath),
		},
		{
			options.NewTerragruntOptionsForTest("/fake/path"),
			"",
			parentTerragruntConfigNotFound("/fake/path/" + DefaultTerragruntConfigPath),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.terragruntOptions.WorkingDir, func(t *testing.T) {
			context := resolveContext{include: mockDefaultInclude, options: testCase.terragruntOptions}
			actualPath, actualErr := context.findInParentFolders()
			if testCase.expectedErr != nil {
				assert.EqualError(t, actualErr, testCase.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
			assert.Equal(t, testCase.expectedPath, actualPath)
		})
	}
}
