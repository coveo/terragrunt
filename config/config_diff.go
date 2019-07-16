package config

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/coveooss/gotemplate/v3/hcl"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Diff find the strings that were removed from the initial config file
func Diff(configString string, terragruntConfig *TerragruntConfig) (string, error) {
	var re = regexp.MustCompile(`".*" {`)
	strippedConfigString := re.ReplaceAllString(configString, "{")
	strippedConfigString = strings.ReplaceAll(strippedConfigString, `"${get_terraform_commands_that_need_vars()}"`, `"apply","console","destroy","import","plan","push","refresh","validate"`)

	marshalConfig, err := hcl.Marshal(terragruntConfig)
	if err != nil {
		return "", err
	}

	var in interface{}
	var out interface{}

	if err := hcl.Unmarshal([]byte(strippedConfigString), &in); err != nil {
		return "", err
	}
	if err := hcl.Unmarshal(marshalConfig, &out); err != nil {
		return "", err
	}

	prettyInputConfig := in.(hcl.Dictionary)["terragrunt"].(hcl.Dictionary).PrettyPrint()
	prettyOutputConfig := out.(hcl.Dictionary).PrettyPrint()

	if strings.Compare(prettyInputConfig, prettyOutputConfig) == 0 {
		return "", nil
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(prettyInputConfig, prettyOutputConfig, false)

	var buffer bytes.Buffer
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffDelete { // Filter deletions
			buffer.WriteString(d.Text)
		}
	}

	result := buffer.String()
	return result, nil
}
