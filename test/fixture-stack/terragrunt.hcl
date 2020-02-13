# Configure Terragrunt to automatically store tfstate files in an S3 bucket
remote_state {
  backend = "s3"

  config = {
    encrypt        = true
    bucket         = "__FILL_IN_BUCKET_NAME__"
    key            = "${path_relative_to_include()}/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "__FILL_IN_LOCK_TABLE_NAME__"
  }
}

inputs = {
  terraform_remote_state_s3_bucket = "__FILL_IN_BUCKET_NAME__"
}

export_variables {
  path = "terraform.tfvars"
}