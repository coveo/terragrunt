terraform {
  source = "github.com/gruntwork-io/terragrunt.git//test/fixture-download/relative?ref=v0.9.9"
}
inputs = {
  name = "World"
}