module github.com/coveooss/terragrunt/v2

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.7.0
	github.com/aws/aws-sdk-go-v2/config v1.4.1
	github.com/aws/aws-sdk-go-v2/credentials v1.3.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.7.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.5.0
	github.com/cheekybits/genny v1.0.0
	github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/coveooss/gotemplate/v3 v3.7.0
	github.com/coveooss/multilogger v0.5.2
	github.com/fatih/color v1.12.0
	github.com/go-errors/errors v1.4.0
	github.com/hashicorp/go-getter v1.5.5
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/hashicorp/terraform v0.15.3
	github.com/matryer/try v0.0.0-20161228173917-9ac251b645a2 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/rs/xid v1.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
	github.com/zclconf/go-cty v1.9.0
	golang.org/x/tools v0.1.4 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/matryer/try.v1 v1.0.0-20150601225556-312d2599e12e
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/hashicorp/go-getter => github.com/coveord/go-getter v1.5.12
	// sum DB is messed up for v2.8.12
	github.com/hashicorp/hcl/v2 => github.com/coveord/hcl/v2 v2.8.102
)
