package config

import (
	"testing"

	"github.com/gruntwork-io/terragrunt/remote"
)

func TestDiff(t *testing.T) {
	type args struct {
		configString     string
		terragruntConfig *TerragruntConfig
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid config",
			args: args{
				configString: `
				terragrunt = {
				  remote_state {
					backend = "s3"
				  }
				}
				`,
				terragruntConfig: &TerragruntConfig{
					RemoteState: &remote.RemoteState{
						Backend: "s3",
					},
				},
			},
			want:    "",
			wantErr: false,
		},

		{
			name: "Typo",
			args: args{
				configString: `
				terragrunt = {
				  descriptio = "Typo"
				}
				`,
				terragruntConfig: &TerragruntConfig{},
			},
			want:    `descriptio = "Typo"`,
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Diff(tt.args.configString, tt.args.terragruntConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Diff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Diff() = %v, want %v", got, tt.want)
			}
		})
	}
}
