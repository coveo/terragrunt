terragrunt = {
  import_variables "test" {
    required_var_files = [
      "vars.json",
    ]

    output_variables_file = "test.tf"
  }
}
