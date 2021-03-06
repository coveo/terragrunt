package remote

import (
	"strings"
	"testing"

	"github.com/coveooss/terragrunt/v2/options"
	"github.com/stretchr/testify/assert"
)

func TestToTerraformInitArgs(t *testing.T) {
	t.Parallel()

	remoteState := State{
		Backend: "s3",
		Config: map[string]interface{}{
			"encrypt": true,
			"bucket":  "my-bucket",
			"key":     "terraform.tfstate",
			"region":  "us-east-1",
		},
	}
	args := remoteState.ToTerraformInitArgs()

	assertTerraformInitArgsEqual(t, args, "-backend-config=encrypt=true -backend-config=bucket=my-bucket -backend-config=key=terraform.tfstate -backend-config=region=us-east-1 -force-copy")
}

func TestToTerraformInitArgsNoBackendConfigs(t *testing.T) {
	t.Parallel()

	remoteState := State{Backend: "s3"}
	args := remoteState.ToTerraformInitArgs()
	assertTerraformInitArgsEqual(t, args, "-force-copy")
}

func TestShouldOverrideExistingRemoteState(t *testing.T) {
	t.Parallel()

	terragruntOptions := options.NewTerragruntOptionsForTest("remote_state_test")

	testCases := []struct {
		existingBackend terraformBackend
		stateFromConfig State
		shouldOverride  bool
	}{
		{terraformBackend{}, State{}, false},
		{terraformBackend{Type: "s3"}, State{Backend: "s3"}, false},
		{terraformBackend{Type: "s3"}, State{Backend: "atlas"}, true},
		{
			terraformBackend{
				Type:   "s3",
				Config: map[string]interface{}{"bucket": "foo", "key": "bar", "region": "us-east-1"},
			},
			State{
				Backend: "s3",
				Config:  map[string]interface{}{"bucket": "foo", "key": "bar", "region": "us-east-1"},
			},
			false,
		}, {
			terraformBackend{
				Type:   "s3",
				Config: map[string]interface{}{"bucket": "foo", "key": "bar", "region": "us-east-1"},
			},
			State{
				Backend: "s3",
				Config:  map[string]interface{}{"bucket": "different", "key": "bar", "region": "us-east-1"},
			},
			true,
		}, {
			terraformBackend{
				Type:   "s3",
				Config: map[string]interface{}{"bucket": "foo", "key": "bar", "region": "us-east-1"},
			},
			State{
				Backend: "s3",
				Config:  map[string]interface{}{"bucket": "foo", "key": "different", "region": "us-east-1"},
			},
			true,
		}, {
			terraformBackend{
				Type:   "s3",
				Config: map[string]interface{}{"bucket": "foo", "key": "bar", "region": "us-east-1"},
			},
			State{
				Backend: "s3",
				Config:  map[string]interface{}{"bucket": "foo", "key": "bar", "region": "different"},
			},
			true,
		},
	}

	for _, testCase := range testCases {
		shouldOverride, err := shouldOverrideExistingRemoteState(&testCase.existingBackend, testCase.stateFromConfig, terragruntOptions)
		assert.Nil(t, err, "Unexpected error: %v", err)
		assert.Equal(t, testCase.shouldOverride, shouldOverride, "Expect shouldOverrideExistingRemoteState to return %t but got %t for existingRemoteState %v and remoteStateFromTerragruntConfig %v", testCase.shouldOverride, shouldOverride, testCase.existingBackend, testCase.stateFromConfig)
	}
}

func assertTerraformInitArgsEqual(t *testing.T, actualArgs []string, expectedArgs string) {
	expected := strings.Split(expectedArgs, " ")
	assert.Len(t, actualArgs, len(expected))

	for _, expectedArg := range expected {
		assert.Contains(t, actualArgs, expectedArg)
	}
}
