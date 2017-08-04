# Terragrunt

Terragrunt is a thin wrapper for [Terraform](https://www.terraform.io/) that provides extra tools for keeping your Terraform configurations DRY, working with multiple Terraform modules, and managing remote state.


## Quick start

1. [Install Terraform](https://www.terraform.io/intro/getting-started/install.html).

1. Install Terragrunt by going to the [Releases Page](https://github.com/gruntwork-io/terragrunt/releases),
   downloading the binary for your OS, renaming it to `terragrunt`, and adding it to your PATH.
     * See the [Install Terragrunt](#install-terragrunt) docs for other installation options.

1. Go into a folder with your Terraform configurations (`.tf` files) and create a `terraform.tfvars` file with a
   `terragrunt = { ... }` block that contains the configuration for Terragrunt (check out the [Use cases](#use-cases)
   section for the types of configuration Terragrunt supports):

    ```hcl
    terragrunt = {
      # (put your Terragrunt configuration here)
    }
    ```

1. Now, instead of running `terraform` directly, run all the standard Terraform commands using `terragrunt`:

    ```bash
    terragrunt get
    terragrunt plan
    terragrunt apply
    terragrunt output
    terragrunt destroy
    ```

   Terragrunt forwards almost all commands, arguments, and options directly to Terraform, using whatever version of
   Terraform you already have installed. However, based on the settings in your `terraform.tfvars` file, Terragrunt can
   configure remote state, locking, extra arguments, and lots more.

1. Terragrunt is a direct implementation of the ideas expressed in 
   [Terraform: Up & Running](http://www.terraformupandrunning.com). Additional background reading that will help
   explain the motivation for Terragrunt includes the Gruntwork.io blog posts
   [How to create reusable infrastructure with Terraform modules](https://blog.gruntwork.io/how-to-create-reusable-infrastructure-with-terraform-modules-25526d65f73d)
   and [How to use Terraform as a team](https://blog.gruntwork.io/how-to-use-terraform-as-a-team-251bc1104973).

1. Check out the [terragrunt-infrastructure-modules-example](https://github.com/gruntwork-io/terragrunt-infrastructure-modules-example)
   and [terragrunt-infrastructure-live-example](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example) 
   repos for fully-working sample code that demonstrates how to use Terragrunt.
   

## Table of Contents

1. [Install Terragrunt](#install-terragrunt)
1. [Use cases](#use-cases)
   1. [Keep your Terraform code DRY](#keep-your-terraform-code-dry)
   1. [Keep your remote state configuration DRY](#keep-your-remote-state-configuration-dry)
   1. [Keep your CLI flags DRY](#keep-your-cli-flags-dry)
   1. [Execute Terraform commands on multiple modules at once](#execute-terraform-commands-on-multiple-modules-at-once)
   1. [Assume a different AWS IAM role to execute Terraform commands](#assume-aws-iam-role)
   1. [Define extra commands](#define-extra-commands)
1. [Terragrunt details](#terragrunt-details)
   1. [AWS credentials](#aws-credentials)
   1. [AWS IAM policies](#aws-iam-policies)
   1. [Interpolation Syntax](#interpolation-syntax)
   1. [CLI options](#cli-options)
   1. [Configuration](#configuration)
   1. [Migrating from Terragrunt v0.11.x and Terraform 0.8.x and older](#migrating-from-terragrunt-v011x-and-terraform-08x-and-older)
   1. [Developing Terragrunt](#developing-terragrunt)
   1. [License](#license)


## Install Terragrunt

Note that third-party Terragrunt packages may not be updated with the latest version, but are often close.
Please check your version against the latest available on the 
[Releases Page](https://github.com/gruntwork-io/terragrunt/releases).

### OSX
You can install Terragrunt on OSX using [Homebrew](https://brew.sh/): `brew install terragrunt`.

### Linux

**WARNING**: the snap installer seems to have a bug where it does not allow Terragrunt to work with Terraform and Git dependencies, so we currently do not recommend using it. See the manual install instructions below, instead.

You can install Terragrunt on Linux systems using [snap](https://snapcraft.io/docs/core/install): `snap install terragrunt`.

### Manual
You can install Terragrunt manually by going to the [Releases Page](https://github.com/gruntwork-io/terragrunt/releases),
downloading the binary for your OS, renaming it to `terragrunt`, and adding it to your PATH.


## Use cases

Terragrunt supports the following use cases:

1. [Keep your Terraform code DRY](#keep-your-terraform-code-dry)
1. [Keep your remote state configuration DRY](#keep-your-remote-state-configuration-dry)
1. [Keep your CLI flags DRY](#keep-your-cli-flags-dry)
1. [Execute Terraform commands on multiple modules at once](#execute-terraform-commands-on-multiple-modules-at-once)


### Keep your Terraform code DRY

* [Motivation](#motivation)
* [Remote Terraform configurations](#remote-terraform-configurations)
* [How to use remote configurations](#how-to-use-remote-configurations)
* [Achieve DRY Terraform code and immutable infrastructure](#achieve-dry-terraform-code-and-immutable-infrastructure)
* [Working locally](#working-locally)
* [Important gotcha: working with relative file paths](#important-gotcha-working-with-relative-file-paths)
* [Using Terragrunt with private Git repos](#using-terragrunt-with-private-git-repos)


#### Motivation

Consider the following file structure, which defines three environments (prod, qa, stage) with the same infrastructure
in each one (an app, a MySQL database, and a VPC):

```
└── live
    ├── prod
    │   ├── app
    │   │   └── main.tf
    │   ├── mysql
    │   │   └── main.tf
    │   └── vpc
    │       └── main.tf
    ├── qa
    │   ├── app
    │   │   └── main.tf
    │   ├── mysql
    │   │   └── main.tf
    │   └── vpc
    │       └── main.tf
    └── stage
        ├── app
        │   └── main.tf
        ├── mysql
        │   └── main.tf
        └── vpc
            └── main.tf
```

The contents of each environment will be more or less identical, except perhaps for a few settings (e.g. the prod
environment may run bigger or more servers). As the size of the infrastructure grows, having to maintain all of this
duplicated code between environments becomes more error prone. You can reduce the amount of copy paste using
[Terraform modules](https://blog.gruntwork.io/how-to-create-reusable-infrastructure-with-terraform-modules-25526d65f73d),
but even the code to instantiate a module and set up input variables, output variables, providers, and remote state
can still create a lot of maintenance overhead.

How can you keep your Terraform code [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) so that you only
have to define it once, no matter how many environments you have?


#### Remote Terraform configurations

Terragrunt has the ability to download remote Terraform configurations. The idea is that you define the Terraform code
for your infrastructure just once, in a single repo, called, for example, `modules`:

```
└── modules
    ├── app
    │   └── main.tf
    ├── mysql
    │   └── main.tf
    └── vpc
        └── main.tf
```

This repo contains typical Terraform code, with one difference: anything in your code that should be different between
environments should be exposed as an input variable. For example, the `app` module might expose the following
variables:

```hcl
variable "instance_count" {
  description = "How many servers to run"
}

variable "instance_type" {
  description = "What kind of servers to run (e.g. t2.large)"
}
```

These variables allow you to run smaller/fewer servers in qa and stage to save money and larger/more servers in prod to
ensure availability and scalability.

In a separate repo, called, for example, `live`, you define the code for all of your environments, which now consists
of just one `.tfvars` file per component (e.g. `app/terraform.tfvars`, `mysql/terraform.tfvars`, etc). This gives you
the following file layout:

```
└── live
    ├── prod
    │   ├── app
    │   │   └── terraform.tfvars
    │   ├── mysql
    │   │   └── terraform.tfvars
    │   └── vpc
    │       └── terraform.tfvars
    ├── qa
    │   ├── app
    │   │   └── terraform.tfvars
    │   ├── mysql
    │   │   └── terraform.tfvars
    │   └── vpc
    │       └── terraform.tfvars
    └── stage
        ├── app
        │   └── terraform.tfvars
        ├── mysql
        │   └── terraform.tfvars
        └── vpc
            └── terraform.tfvars
```

Notice how there are no Terraform configurations (`.tf` files) in any of the folders. Instead, each `.tfvars` file
specifies a `terraform { ... }` block that specifies from where to download the Terraform code, as well as the
environment-specific values for the input variables in that Terraform code. For example,
`stage/app/terraform.tfvars` may look like this:

```hcl
terragrunt = {
  terraform {
    source = "git::git@github.com:foo/modules.git//app?ref=v0.0.3"
  }
}

instance_count = 3
instance_type = "t2.micro"
```

*(Note: the double slash (`//`) is intentional and required. It's part of Terraform's Git syntax for [module
sources](https://www.terraform.io/docs/modules/sources.html).)*

And `prod/app/terraform.tfvars` may look like this:

```hcl
terragrunt = {
  terraform {
    source = "git::git@github.com:foo/modules.git//app?ref=v0.0.1"
  }
}

instance_count = 10
instance_type = "m2.large"
```

Notice how the two `terraform.tfvars` files set the `source` URL to the same `app` module, but at different
versions (i.e. `stage` is testing out a newer version of the module). They also set the parameters for the
`app` module to different values that are appropriate for the environment: smaller/fewer servers in `stage`
to save money, larger/more instances in `prod` for scalability and high availability.

Check out the [terragrunt-infrastructure-modules-example](https://github.com/gruntwork-io/terragrunt-infrastructure-modules-example)
and [terragrunt-infrastructure-live-example](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example) 
repos for fully-working sample code that demonstrates this new folder structure.




#### How to use remote configurations

Once you've set up your `live` and `modules` repositories, all you need to do is run `terragrunt` commands in the
`live` repository. For example, to deploy the `app` module in qa, you would do the following:

```
cd live/qa/app
terragrunt apply
```

When Terragrunt finds the `terraform` block with a `source` parameter in `live/qa/app/terraform.tfvars` file, it will:

1. Download the configurations specified via the `source` parameter into a temporary folder. This downloading is done
   by using the [terraform init command](https://www.terraform.io/docs/commands/init.html), so the `source` parameter
   supports the exact same syntax as the [module source](https://www.terraform.io/docs/modules/sources.html) parameter,
   including local file paths, Git URLs, and Git URLs with `ref` parameters (useful for checking out a specific tag,
   commit, or branch of Git repo). Terragrunt will download all the code in the repo (i.e. the part before the
   double-slash `//`) so that relative paths work correctly between modules in that repo.

1. Copy all files from the current working directory into the temporary folder. This way, Terraform will automatically
   read in the variables defined in the `terraform.tfvars` file.

1. Execute whatever Terraform command you specified in that temporary folder.


#### Achieve DRY Terraform code and immutable infrastructure

With this new approach, copy/paste between environments is minimized. The `.tfvars` files contain solely the variables
that are different between environments. To create a new environment, you copy an old one and update just the
environment-specific values in the `.tfvars` files, which is about as close to the "essential complexity" of the
problem as you can get.

Just as importantly, since the Terraform module code is now defined in a single repo, you can version it (e.g., using Git
tags and referencing them using the `ref` parameter in the `source` URL, as in the `stage/app/terraform.tfvars` and
`prod/app/terraform.tfvars` examples above), and promote a single, immutable version through each environment (e.g.,
qa -> stage -> prod). This idea is inspired by Kief Morris' blog post [Using Pipelines to Manage Environments with
Infrastructure as Code](https://medium.com/@kief/https-medium-com-kief-using-pipelines-to-manage-environments-with-infrastructure-as-code-b37285a1cbf5).


#### Working locally

If you're testing changes to a local copy of the `modules` repo, you you can use the `--terragrunt-source` command-line
option or the `TERRAGRUNT_SOURCE` environment variable to override the `source` parameter. This is useful to point
Terragrunt at a local checkout of your code so you can do rapid, iterative, make-a-change-and-rerun development:

```
cd live/stage/app
terragrunt apply --terragrunt-source ../../../modules//app
```

*(Note: the double slash (`//`) here too is intentional and required. Terragrunt downloads all the code in the folder
before the double-slash into the temporary folder so that relative paths between modules work correctly.)*


#### Important gotcha: working with relative file paths

One of the gotchas with downloading Terraform configurations is that when you run `terragrunt apply` in folder `foo`,
Terraform will actually execute in some temporary folder such as `/tmp/foo`. That means you have to be especially
careful with relative file paths, as they will be relative to that temporary folder and not the folder where you ran
Terragrunt!

In particular:

* **Command line**: When using file paths on the command line, such as passing an extra `-var-file` argument, you
  should use absolute paths:

    ```bash
    # Use absolute file paths on the CLI!
    terragrunt apply -var-file /foo/bar/extra.tfvars
    ```

* **Terragrunt configuration**: When using file paths directly in your Terragrunt configuration (`terraform.tfvars`),
  such as in an `extra_arguments` block, you can't use hard-coded absolute file paths, or it won't work on your
  teammates' computers. Therefore, you should utilize the Terragrunt built-in function `get_tfvars_dir()` to use
  a relative file path:

    ```hcl
    terragrunt = {
      terraform {
        source = "git::git@github.com:foo/modules.git//frontend-app?ref=v0.0.3"

        extra_arguments "custom_vars" {
          commands = [
            "apply",
            "plan",
            "import",
            "push",
            "refresh"
          ]

          # With the get_tfvars_dir() function, you can use relative paths!
          arguments = [
            "-var-file=${get_tfvars_dir()}/../common.tfvars",
            "-var-file=terraform.tfvars"
          ]
        }
      }
    }
    ```

  See the [get_tfvars_dir()](#get_tfvars_dir) documentation for more details.


#### Using Terragrunt with private Git repos

The easiest way to use Terragrunt with private Git repos is to use SSH authentication. 
Configure your Git account so you can use it with SSH 
(see the [guide for GitHub here](https://help.github.com/articles/connecting-to-github-with-ssh/))
and use the SSH URL for your repo, prepended with `git::ssh://`: 

```hcl
terragrunt = {
  terraform {
    source = "git::ssh://git@github.com/foo/modules.git//path/to/module?ref=v0.0.1"
  }
}
```
Look up the Git repo for your repository to find the proper format. 

Note: In automated pipelines, you may need to run the following command for your 
Git repository prior to calling `terragrunt` to ensure that the ssh host is registered 
locally, e.g.:

```
$ ssh -T -oStrictHostKeyChecking=no git@github.com || true
```


### Keep your remote state configuration DRY

* [Motivation](#motivation-1)
* [Filling in remote state settings with Terragrunt](#filling-in-remote-state-settings-with-terragrunt)
* [Create remote state and locking resources automatically](#create-remote-state-and-locking-resources-automatically)


#### Motivation

Terraform supports [remote state storage](https://www.terraform.io/docs/state/remote.html) via a variety of
[backends](https://www.terraform.io/docs/backends) that you configure as follows:

```hcl
terraform {
  backend "s3" {
    bucket     = "my-terraform-state"
    key        = "frontend-app/terraform.tfstate"
    region     = "us-east-1"
    encrypt    = true
    lock_table = "my-lock-table"
  }
}
```

Unfortunately, the `backend` configuration does not support interpolation. This makes it hard to keep your code
[DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) if you have multiple Terraform modules. For example,
consider the following folder structure, which uses different Terraform modules to deploy a backend app, frontend app,
MySQL database, and a VPC:

```
├── backend-app
│   └── main.tf
├── frontend-app
│   └── main.tf
├── mysql
│   └── main.tf
└── vpc
    └── main.tf
```

To use remote state with each of these modules, you would have to copy/paste the exact same `backend` configuration
into each of the `main.tf` files. The only thing that would differ between the configurations would be the `key`
parameter: e.g., the `key` for `mysql/main.tf` might be `mysql/terraform.tfstate` and the `key` for
`frontend-app/main.tf` might be `frontend-app/terraform.tfstate`.

To keep your remote state configuration DRY, you can use Terragrunt. You still have to specify the `backend` you want
to use in each module, but instead of copying and pasting the configuration settings over and over again into each
`main.tf` file, you can leave them blank:

```hcl
terraform {
  # The configuration for this backend will be filled in by Terragrunt
  backend "s3" {}
}
```


#### Filling in remote state settings with Terragrunt

To fill in the settings via Terragrunt, create a `terraform.tfvars` file in the root folder and in each of the
Terraform modules:

```
├── terraform.tfvars
├── backend-app
│   ├── main.tf
│   └── terraform.tfvars
├── frontend-app
│   ├── main.tf
│   └── terraform.tfvars
├── mysql
│   ├── main.tf
│   └── terraform.tfvars
└── vpc
    ├── main.tf
    └── terraform.tfvars
```

In your **root** `terraform.tfvars` file, you can define your entire remote state configuration just once in a
`remote_state` block, as follows:

```hcl
terragrunt = {
  remote_state {
    backend = "s3"
    config {
      bucket     = "my-terraform-state"
      key        = "${path_relative_to_include()}/terraform.tfstate"
      region     = "us-east-1"
      encrypt    = true
      lock_table = "my-lock-table"
    }
  }
}
```

The `remote_state` block supports all the same [backend types](https://www.terraform.io/docs/backends/types/index.html)
as Terraform. The next time you run `terragrunt`, it will automatically configure all the settings in the
`remote_state.config` block, if they aren't configured already, by calling [terraform
init](https://www.terraform.io/docs/commands/init.html).

In each of the **child** `terraform.tfvars` files, such as `mysql/terraform.tfvars`, you can tell Terragrunt to
automatically include all the settings from the root `terraform.tfvars` file as follows:

```hcl
terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }
}
```

The `include` block tells Terragrunt to use the exact same Terragrunt configuration from the `terraform.tfvars` file
specified via the `path` parameter. It behaves exactly as if you had copy/pasted the Terraform configuration from 
the root `terraform.tfvars` file into `mysql/terraform.tfvars`, but this approach is much easier to maintain!

The child `.tfvars` file's `terragrunt.terraform` settings will be merged into the parent file's `terragrunt.terraform`
settings as follows:

* If an `extra_arguments` block in the child has the same name as an `extra_arguments` block in the parent,
  then the child's block will override the parent's.
  * Specifying an empty `extra_arguments` block in a child with the same name will effectively remove the parent's block.
* If an `extra_arguments` block in the child has a different name than `extra_arguments` blocks in the parent,
  then both the parent and child's `extra_arguments` will be effective.
  * The child's `extra_arguments` will be placed _after_ the parent's `extra_arguments` on the terraform command line.
  * Therefore, if a child's and parent's `extra_arguments` include `.tfvars` files with the same variable defined,
    the value from the `.tfvars` file from the child's `extra_arguments` will be used by terraform.
* The `source` field in the child will override `source` field in the parent

Other settings in the child `.tfvars` file's `terragrunt` block (e.g. `remote_state`) override the respective
settings in the parent.

The `terraform.tfvars` files above use two Terragrunt built-in functions:

* `find_in_parent_folders()`: This function returns the path to the first `terraform.tfvars` file it finds in the parent
  folders above the current `terraform.tfvars` file. In the example above, the call to `find_in_parent_folders()` in
  `mysql/terraform.tfvars` will return `../terraform.tfvars`. This way, you don't have to hard code the `path`
  parameter in every module.

* `path_relative_to_include()`: This function returns the relative path between the current `terraform.tfvars`
  file and the path specified in its `include` block. We typically use this in a root `terraform.tfvars` file so that
  each Terraform child module stores its Terraform state at a different `key`. For example, the `mysql` module will
  have its `key` parameter resolve to `mysql/terraform.tfstate` and the `frontend-app` module will have its `key`
  parameter resolve to `frontend-app/terraform.tfstate`.

See [the Interpolation Syntax docs](#interpolation-syntax) for more info.

Check out the [terragrunt-infrastructure-modules-example](https://github.com/gruntwork-io/terragrunt-infrastructure-modules-example)
and [terragrunt-infrastructure-live-example](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example) 
repos for fully-working sample code that demonstrates how to use Terragrunt to manage remote state.




#### Create remote state and locking resources automatically

When you run `terragrunt` with `remote_state` configuration, it will automatically create the following resources if
they don't already exist:

* **S3 bucket**: If you are using the [S3 backend](https://www.terraform.io/docs/backends/types/s3.html) for remote
  state storage and the `bucket` you specify in `remote_state.config` doesn't already exist, Terragrunt will create it
  automatically, with [versioning enabled](http://docs.aws.amazon.com/AmazonS3/latest/dev/Versioning.html).

* **DynamoDB table**: If you are using the [S3 backend](https://www.terraform.io/docs/backends/types/s3.html) for
  remote state storage and you specify a `lock_table` (a [DynamoDB table used for
  locking](https://www.terraform.io/docs/backends/types/s3.html#lock_table)) in `remote_state.config`, if that table
  doesn't already exist, Terragrunt will create it automatically, including a primary key called `LockID`.

**Note**: If you specify a `profile` key in `remote_state.config`, Terragrunt will automatically use this AWS profile
when creating the S3 bucket or DynamoDB table.


### Keep your CLI flags DRY

* [Motivation](#motivation-2)
* [Multiple extra_arguments blocks](#multiple-extra_arguments-blocks)
* [Required and optional var-files](#required-and-optional-var-files)
* [Handling whitespace](#handling-whitespace)

#### Motivation

Sometimes you may need to pass extra CLI arguments every time you run certain `terraform` commands. For example, you
may want to set the `lock-timeout` setting to 20 minutes for all commands that may modify remote state so that
Terraform will keep trying to acquire a lock for up to 20 minutes if someone else already has the lock rather than
immediately exiting with an error.

You can configure Terragrunt to pass specific CLI arguments for specific commands using an `extra_arguments` block
in your `terraform.tfvars` file:

```hcl
terragrunt = {
  terraform {
    # Force Terraform to keep trying to acquire a lock for
    # up to 20 minutes if someone else already has the lock
    extra_arguments "retry_lock" {
      commands = [
        "init",
        "apply",
        "refresh",
        "import",
        "plan",
        "taint",
        "untaint"
      ]

      arguments = [
        "-lock-timeout=20m"
      ]
    }
  }
}
```

Each `extra_arguments` block includes an arbitrary name (in the example above, `retry_lock`), a list of `commands` to
which the extra arguments should be add, a list of `arguments` or `required_var_files` or `optional_var_files` to add.
With the configuration above, when you run `terragrunt apply`, Terragrunt will call Terraform as follows:

When available, it is preferable to use interpolation functions such as
[get_terraform_commands_that_need_locking](#get_terraform_commands_that_need_locking) and
[get_terraform_commands_that_need_vars](#get_terraform_commands_that_need_vars)
since they provide the complete list of terraform commands that make use of the desired parameter:

```hcl
terragrunt = {
  terraform {
    # Force Terraform to keep trying to acquire a lock for up to 20 minutes if someone else already has the lock
    extra_arguments "retry_lock" {
      commands  = ["${get_terraform_commands_that_need_locking()}"]
      arguments = ["-lock-timeout=20m"]
    }
  }
}
```

```
> terragrunt apply

terraform apply -lock-timeout=20m
```


#### Multiple extra_arguments blocks

You can specify one or more `extra_arguments` blocks. The `arguments` in each block will be applied any time you call
`terragrunt` with one of the commands in the `commands` list. If more than one `extra_arguments` block matches a
command, the arguments will be added in the order of of appearance in the configuration. For example, in addition to
lock settings, you may also want to pass custom `-var-file` arguments to several commands:

```hcl
terragrunt = {
  terraform {
    # Force Terraform to keep trying to acquire a lock for
    # up to 20 minutes if someone else already has the lock
    extra_arguments "retry_lock" {
      commands = [
        "init",
        "apply",
        "refresh",
        "import",
        "plan",
        "taint",
        "untaint"
      ]

      arguments = [
        "-lock-timeout=20m"
      ]
    }

    # Pass custom var files to Terraform
    extra_arguments "custom_vars" {
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]

      arguments = [
        "-var", "foo=bar",
        "-var", "region=us-west-1"
      ]
    }
  }
}
```

With the configuration above, when you run `terragrunt apply`, Terragrunt will call Terraform as follows:

```
> terragrunt apply

terraform apply -lock-timeout=20m -var foo=bar -var region=us-west-1
```

#### Required and optional var-files

One common usage of extra_arguments is to include tfvars files. instead of using arguments, it is simpler to use either `required_var_files`
or `optional_var_files`. Both options require only to provide the list of file to include. The only difference is that `required_var_files`
will add the extra argument `-var-file=<your file>` for each file specified and if they don't exist, terraform will complain. Using
`optional_var_files` instead, terragrunt will only add the `-var-file=<your file>` for existing files. This allows many conditional
configurations based on environment variables as you can see in the following example:

```
/my/tf
├── terraform.tfvars
├── prod.tfvars
├── us-west-2.tfvars
├── backend-app
│   ├── main.tf
│   ├── dev.tfvars
│   └── terraform.tfvars
├── frontend-app
│   ├── main.tf
│   ├── us-east-1.tfvars
│   └── terraform.tfvars
```

```hcl
terragrunt = {
  terraform {
    extra_arguments "conditional_vars" {
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]

      required_var_files = [
        "${get_parent_tfvars_dir()}/terraform.tfvars"
      ]

      optional_var_files = [
        "${get_parent_tfvars_dir()}/${get_env("TF_VAR_env", "dev")}.tfvars",
        "${get_parent_tfvars_dir()}/${get_env("TF_VAR_region", "us-east-1")}.tfvars"
        "${get_tfvars_dir()}/${get_env("TF_VAR_env", "dev")}.tfvars",
        "${get_tfvars_dir()}/${get_env("TF_VAR_region", "us-east-1")}.tfvars"
      ]
    }
  }
```

See the [get_tfvars_dir()](#get_tfvars_dir) and [get_parent_tfvars_dir()](#get_parent_tfvars_dir) documentation for more details.

_Note that terragrunt cannot interpolate terraform variables (${var.xxx}) in the terragrunt configuration,
your variables have to be defined through TF_VAR_xxx environment variable to be referred by terragrunt._

With the configuration above, when you run `terragrunt apply-all`, Terragrunt will call Terraform as follows:

```
> terragrunt apply-all
[backend-app]  terraform apply -var-file=/my/tf/terraform.tfvars -var-file=/my/tf/backend-app/dev.tfvars
[frontend-app] terraform apply -var-file=/my/tf/terraform.tfvars -var-file=/my/tf/frontend-app/us-east-1.tfvars

> TF_VAR_env=prod terragrunt apply-all
[backend-app]  terraform apply -var-file=/my/tf/terraform.tfvars -var-file=/my/tf/prod.tfvars
[frontend-app] terraform apply -var-file=/my/tf/terraform.tfvars -var-file=/my/tf/prod.tfvars -var-file=/my/tf/frontend-app/us-east-1.tfvars

> TF_VAR_env=prod TF_VAR_region=us-west-2 terragrunt apply-all
[backend-app]  terraform apply -var-file=/my/tf/terraform.tfvars -var-file=/my/tf/prod.tfvars -var-file=/my/tf/us-west-2.tfvars
[frontend-app] terraform apply -var-file=/my/tf/terraform.tfvars -var-file=/my/tf/prod.tfvars -var-file=/my/tf/us-west-2.tfvars
```

#### Handling whitespace

The list of arguments cannot include whitespaces, so if you need to pass command line arguments that include
spaces (e.g. `-var bucket=example.bucket.name`), then each of the arguments will need to be a separate item in the
`arguments` list:

```hcl
terragrunt = {
  terraform {
    extra_arguments "bucket" {
      arguments = [
        "-var", "bucket=example.bucket.name",
      ]
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]
    }
  }
}
```

With the configuration above, when you run `terragrunt apply`, Terragrunt will call Terraform as follows:

```
> terragrunt apply

terraform apply -var bucket=example.bucket.name
```


### Execute Terraform commands on multiple modules at once

* [Motivation](#motivation-3)
* [The apply-all, destroy-all, output-all and plan-all commands](#the-apply-all-destroy-all-output-all-and-plan-all-commands)
* [Dependencies between modules](#dependencies-between-modules)


#### Motivation

Let's say your infrastructure is defined across multiple Terraform modules:

```
root
├── backend-app
│   └── main.tf
├── frontend-app
│   └── main.tf
├── mysql
│   └── main.tf
├── redis
│   └── main.tf
└── vpc
    └── main.tf
```

There is one module to deploy a frontend-app, another to deploy a backend-app, another for the MySQL database, and so
on. To deploy such an environment, you'd have to manually run `terraform apply` in each of the subfolder, wait for it
to complete, and then run `terraform apply` in the next subfolder. How do you avoid this tedious and time-consuming
process?


#### The apply-all, destroy-all, output-all and plan-all commands

To be able to deploy multiple Terraform modules in a single command, add a `terraform.tfvars` file to each module:

```
root
├── backend-app
│   ├── main.tf
│   └── terraform.tfvars
├── frontend-app
│   ├── main.tf
│   └── terraform.tfvars
├── mysql
│   ├── main.tf
│   └── terraform.tfvars
├── redis
│   ├── main.tf
│   └── terraform.tfvars
└── vpc
    ├── main.tf
    └── terraform.tfvars
```

Inside each `terraform.tfvars` file, add a `terragrunt = { ... }` block to identify this as a module managed by
Terragrunt (the block can be empty or include any of the configs described in this documentation):

```hcl
terragrunt = {
  # Put your Terragrunt configuration here
}
```

Now you can go into the `root` folder and deploy all the modules within it by using the `apply-all` command:

```
cd root
terragrunt apply-all
```

When you run this command, Terragrunt will recursively look through all the subfolders of the current working
directory, find all `terraform.tfvars` files with a `terragrunt = { ... }` block, and run `terragrunt apply` in each
one concurrently.

Similarly, to undeploy all the Terraform modules, you can use the `destroy-all` command:

```
cd root
terragrunt destroy-all
```

To see the currently applied outputs of all of the subfolders, you can use the `output-all` command:

```
cd root
terragrunt output-all
```

Finally, if you make some changes to your project, you could evaluate the impact by using `plan-all` command:

Note: It is important to realize that you could get errors running `plan-all` if you have dependencies between your projects
and some of those dependencies haven't been applied yet.

_Ex: If module A depends on module B and module B hasn't been applied yet, then plan-all will show the plan for B,
but exit with an error when trying to show the plan for A._

```
cd root
terragrunt plan-all
```

If your modules have dependencies between them—for example, you can't deploy the backend-app until MySQL and redis are
deployed—you'll need to express those dependencies in your Terragrunt configuration as explained in the next section.


#### Dependencies between modules

Consider the following file structure:

```
root
├── backend-app
│   ├── main.tf
│   └── terraform.tfvars
├── frontend-app
│   ├── main.tf
│   └── terraform.tfvars
├── mysql
│   ├── main.tf
│   └── terraform.tfvars
├── redis
│   ├── main.tf
│   └── terraform.tfvars
└── vpc
    ├── main.tf
    └── terraform.tfvars
```

Let's assume you have the following dependencies between Terraform modules:

* `backend-app` depends on `mysql`, `redis`, and `vpc`
* `frontend-app` depends on `backend-app` and `vpc`
* `mysql` depends on `vpc`
* `redis` depends on `vpc`
* `vpc` has no dependencies

You can express these dependencies in your `terraform.tfvars` config files using a `dependencies` block. For example,
in `backend-app/terraform.tfvars` you would specify:

```hcl
terragrunt = {
  dependencies {
    paths = ["../vpc", "../mysql", "../redis"]
  }
}
```

Similarly, in `frontend-app/terraform.tfvars`, you would specify:

```hcl
terragrunt = {
  dependencies {
    paths = ["../vpc", "../backend-app"]
  }
}
```

Once you've specified the dependencies in each `terraform.tfvars` file, when you run the `terragrunt apply-all` or
`terragrunt destroy-all`, Terragrunt will ensure that the dependencies are applied or destroyed, respectively, in the
correct order. For the example at the start of this section, the order for the `apply-all` command would be:

1. Deploy the VPC
1. Deploy MySQL and Redis in parallel
1. Deploy the backend-app
1. Deploy the frontend-app

If any of the modules fail to deploy, then Terragrunt will not attempt to deploy the modules that depend on them. Once
you've fixed the error, it's usually safe to re-run the `apply-all` or `destroy-all` command again, since it'll be a
no-op for the modules that already deployed successfully, and should only affect the ones that had an error the last
time around.

### Assume AWS IAM role

Terraform already provides the functionality to configure AWS provider that assume a different IAM Role when retrieving and creating AWS resources.
But when we use terragrunt to configure S3 backend to store our remote states, terraform uses the current user rights to access and configure the remote state file and to manage locking operation in the DynamoDB database.

Since the state files may contain secrets, it is often required to restrict access to these files. But event if the AWS provider is configured to allow access to the state file
by assuming a role, the call will fail if the current user does not have a direct access to theses files.

Moreover, if the user has configured its AWS profile (in .aws/config) to assume a role instead of directly using credentials, terraform would not be
able to recognize that configuration and will complain that there is `No valid credential sources found for AWS Provider`

```
[profile deploy]
source_profile = default
role_arn = arn:aws:iam::9999999999999:role/deploy-role
region = us-east-1
```

#### Configure role

To solve that problem, it is possible to tell terragrunt to assume a different IAM role when it calls terraform operations.

```hcl
terragrunt = {
  assume_role = "arn:aws:iam::9999999999999:role/deploy-terraform-role"
}
```

The `assume_role` configuration could be defined in any terragrunt configuration files. If it is defined at several level, the leaf configuration will prevail.

### Define extra commands

Since Terragrunt configure the execution context in temporary folder, it may be useful to execute other command than terraform in that context after
the terraform remote state has been configured.

#### Configure extra commands

```hcl
terragrunt = {
  extra_command "name" {
    commands  = [list of commands]  # optional (default use name as the command)
    os        = [list of os]        # optional (default run on all os, os name are those supported by go, i.e. linux, darwin, windows)
    use_state = true or false       # optional (default = true)
  }
}
```

#### Example of extra commands

```hcl
  # Add extra commands to terragrunt
  extra_command "shell" {
    commands = ["bash", "sh", "zsh", "fish", "ls"]
    os       = ["darwin", "linux"]
  }
```

So the following commands do:

- starts a shell into the temporary folder

```bash
> terragrunt bash
> terragrunt sh
> terragrunt zsh
> terragrunt fish
> terragrunt shell
```

- List the content of the temporary folder

```bash
> terragrunt ls -al
```

The name `shell` used to name the extra_command group could also be used as a command. It acts as an alias for the first command in `commands` list.

## Terragrunt details

This section contains detailed documentation for the following aspects of Terragrunt:

1. [AWS credentials](#aws-credentials)
1. [AWS IAM policies](#aws-iam-policies)
1. [Interpolation Syntax](#interpolation-syntax)
1. [CLI options](#cli-options)
1. [Configuration](#configuration)
1. [Migrating from Terragrunt v0.11.x and Terraform 0.8.x and older](#migrating-from-terragrunt-v011x-and-terraform-08x-and-older)
1. [Developing Terragrunt](#developing-terragrunt)
1. [License](#license)


### AWS credentials

Terragrunt uses the official [AWS SDK for Go](https://aws.amazon.com/sdk-for-go/), which
means that it will automatically load credentials using the 
[AWS standard approach](https://aws.amazon.com/blogs/security/a-new-and-standardized-way-to-manage-credentials-in-the-aws-sdks/). If you need help configuring your credentials, please refer to the [Terraform docs](https://www.terraform.io/docs/providers/aws/#authentication).


### AWS IAM policies

Your AWS user must have an [IAM 
policy](http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/access-control-identity-based.html) 
which grants permissions for interacting with DynamoDB and S3. Terragrunt will automatically create
the configured DynamoDB tables and S3 buckets for storing remote state if they do not already exist.

The following is an example IAM policy for use with Terragrunt. The policy grants the following permissions:

* all DynamoDB permissions in all regions for tables used by Terragrunt
* all S3 permissions for buckets used by Terragrunt

Before using this policy, make sure to replace `1234567890` with your AWS account id and `terragrunt*` with
your organization's naming convention for AWS resources for Terraform remote state.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowAllDynamoDBActionsOnAllTerragruntTables",
            "Effect": "Allow",
            "Action": "dynamodb:*",
            "Resource": [
                "arn:aws:dynamodb:*:1234567890:table/terragrunt*"
            ]
        },
        {
            "Sid": "AllowAllS3ActionsOnTerragruntBuckets",
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": [
                "arn:aws:s3:::terragrunt*",
                "arn:aws:s3:::terragrunt*/*"
            ]
        }
    ]
}
```

### Interpolation syntax

Terragrunt allows you to use [Terraform interpolation syntax](https://www.terraform.io/docs/configuration/interpolation.html)
(`${...}`) to call specific Terragrunt built-in functions. Note that Terragrunt built-in functions **only** work within a 
`terragrunt = { ... }` block. Terraform does NOT process interpolations in `.tfvars` files.

* [find_in_parent_folders()](#find_in_parent_folders)
* [path_relative_to_include()](#path_relative_to_include)
* [path_relative_from_include()](#path_relative_from_include)
* [get_env(NAME, DEFAULT)](#get_env)
* [get_tfvars_dir()](#get_tfvars_dir)
* [get_parent_tfvars_dir()](#get_parent_tfvars_dir)
* [get_terraform_commands_that_need_vars()](#get_terraform_commands_that_need_vars)
* [get_terraform_commands_that_need_input()](#get_terraform_commands_that_need_input)
* [get_terraform_commands_that_need_locking()](#get_terraform_commands_that_need_locking)
* [get_aws_account_id()](#get_aws_account_id)


#### find_in_parent_folders

`find_in_parent_folders()` searches up the directory tree from the current `.tfvars` file and returns the relative path
to to the first `terraform.tfvars` in a parent folder or exit with an error if no such file is found. This is
primarily useful in an `include` block to automatically find the path to a parent `.tfvars` file:

```hcl
terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }
}
```


#### path_relative_to_include

`path_relative_to_include()` returns the relative path between the current `.tfvars` file and the `path` specified in
its `include` block. For example, consider the following folder structure:

```
├── terraform.tfvars
└── prod
    └── mysql
        └── terraform.tfvars
└── stage
    └── mysql
        └── terraform.tfvars
```

Imagine `prod/mysql/terraform.tfvars` and `stage/mysql/terraform.tfvars` include all settings from the root
`terraform.tfvars` file:

```hcl
terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }
}
```

The root `terraform.tfvars` can use the `path_relative_to_include()` in its `remote_state` configuration to ensure
each child stores its remote state at a different `key`:

```hcl
terragrunt = {
  remote_state {
    backend = "s3"
    config {
      bucket = "my-terraform-bucket"
      region = "us-east-1"
      key    = "${path_relative_to_include()}/terraform.tfstate"
    }
  }
}
```

The resulting `key` will be `prod/mysql/terraform.tfstate` for the prod `mysql` module and
`stage/mysql/terraform.tfstate` for the stage `mysql` module.


#### path_relative_from_include

`path_relative_from_include()` returns the relative path between the `path` specified in its `include` block and the current
`.tfvars` file (it is the counterpart of `path_relative_to_include()`). For example, consider the following folder structure:

```
├── sources
|  ├── mysql
|  |  └── *.tf
|  └── secrets
|     └── mysql
|         └── *.tf
└── terragrunt
  └── common.tfvars
  ├── mysql
  |  └── terraform.tfvars
  ├── secrets
  |  └── mysql
  |     └── terraform.tfvars
  └── terraform.tfvars
```

Imagine `terragrunt/mysql/terraform.tfvars` and `terragrunt/secrets/mysql/terraform.tfvars` include all settings from the root
`terraform.tfvars` file:

```hcl
terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }
}
```

The root `terraform.tfvars` can use the `path_relative_from_include()` in combination with `path_relative_to_include()` in its `source` configuration to retrieve the relative terraform source code from the terragrunt configuration file:

```hcl
terragrunt = {
  terraform {
    source = "${path_relative_from_include()}/../sources//${path_relative_to_include()}"
  }
  ...
}
```

The resulting `source` will be `../../sources//mysql` for `mysql` module and `../../../sources//secrets/mysql` for `secrets/mysql` module.

Another use case would be to add extra argument to include the common.tfvars file for all subdirectories:

```hcl
terragrunt = {
  terraform = {
    ...

    extra_arguments "common_var" {
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]

      arguments = [
        "-var-file=${get_tfvars_dir()}/${path_relative_from_include()}/common.tfvars",
      ]
    }
  }
}
```

This allows proper retrieval of the `common.tfvars` from whatever the level of subdirectories we have.


#### get_env

`get_env(NAME, DEFAULT)` returns the value of the environment variable named `NAME` or `DEFAULT` if that environment
variable is not set. Example:

```hcl
terragrunt = {
  remote_state {
    backend = "s3"
    config {
      bucket = "${get_env("BUCKET", "my-terraform-bucket")}"
    }
  }
}
```

Note that [Terraform will read environment
variables](https://www.terraform.io/docs/configuration/environment-variables.html#tf_var_name) that start with the
prefix `TF_VAR_`, so one way to share the a variable named `foo` between Terraform and Terragrunt is to set its value
as the environment variable `TF_VAR_foo` and to read that value in using this `get_env()` built-in function.


#### get_tfvars_dir

`get_tfvars_dir()` returns the directory where the Terragrunt configuration file (by default, `terraform.tfvars`) lives.
This is useful when you need to use relative paths with [remote Terraform
configurations](#remote-terraform-configurations) and you want those paths relative to your Terragrunt configuration
file and not relative to the temporary directory where Terragrunt downloads the code.

For example, imagine you have the following file structure:

```
/terraform-code
├── common.tfvars
├── frontend-app
│   └── terraform.tfvars
```

Inside of `/terraform-code/frontend-app/terraform.tfvars` you might try to write code that looks like this:

```hcl
terragrunt = {
  terraform {
    source = "git::git@github.com:foo/modules.git//frontend-app?ref=v0.0.3"

    extra_arguments "custom_vars" {
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]

      arguments = [
        "-var-file=../common.tfvars", # Note: This relative path will NOT work correctly!
        "-var-file=terraform.tfvars"
      ]
    }
  }
}
```

Note how the `source` parameter is set, so Terragrunt will download the `frontend-app` code from the `modules` repo
into a temporary folder and run `terraform` in that temporary folder. Note also that there is an `extra_arguments`
block that is trying to allow the `frontend-app` to read some shared variables from a `common.tfvars` file.
Unfortunately, the relative path (`../common.tfvars`) won't work, as it will be relative to the temporary folder!
Moreover, you can't use an absolute path, or the code won't work on any of your teammates' computers.

To make the relative path work, you need to use `get_tfvars_dir()` to combine the path with the folder where
the `.tfvars` file lives:

```hcl
terragrunt = {
  terraform {
    source = "git::git@github.com:foo/modules.git//frontend-app?ref=v0.0.3"

    extra_arguments "custom_vars" {
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]

      # With the get_tfvars_dir() function, you can use relative paths!
      arguments = [
        "-var-file=${get_tfvars_dir()}/../common.tfvars",
        "-var-file=terraform.tfvars"
      ]
    }
  }
}
```

For the example above, this path will resolve to `/terraform-code/frontend-app/../common.tfvars`, which is exactly
what you want.


#### get_parent_tfvars_dir

`get_parent_tfvars_dir()` returns the absolute directory where the Terragrunt parent configuration file (by default, `terraform.tfvars`) lives.
This is useful when you need to use relative paths with [remote Terraform configurations](#remote-terraform-configurations) and you want
those paths relative to your parent Terragrunt configuration file and not relative to the temporary directory where Terragrunt downloads
the code.

This function is very similar to [get_tfvars_dir()](#get_tfvars_dir) except it returns the root instead of the leaf of your terragrunt
configuration folder.

```
/terraform-code
├── terraform.tfvars
├── common.tfvars
├── app1
│   └── terraform.tfvars
├── tests
│   ├── app2
│   |   └── terraform.tfvars
│   └── app3
│       └── terraform.tfvars
```

```hcl
terragrunt = {
  terraform {
    extra_arguments "common_vars" {
      commands = [
        "apply",
        "plan",
        "import",
        "push",
        "refresh"
      ]

      arguments = [
        "-var-file=${get_parent_tfvars_dir()}/common.tfvars"
      ]
    }
  }
}
```

The common.tfvars located in the terraform root folder will be included by all applications, whatever their relative location to the root.

#### get_terraform_commands_that_need_vars

`get_terraform_commands_that_need_vars()`

Returns the list of terraform commands that accept -var and -var-file parameters. This function is used when defining [extra_arguments](#keep-your-cli-flags-dry).

```
terragrunt = {
  terraform = {
    ...

    extra_arguments "common_var" {
      commands  = ["${get_terraform_commands_that_need_vars()}"]
      arguments = ["-var-file=${get_aws_account_id()}.tfvars"]
    }
  }
}
```

#### get_terraform_commands_that_need_input

`get_terraform_commands_that_need_input()`

Returns the list of terraform commands that accept -input=(true or false) parameter. This function is used when defining [extra_arguments](#keep-your-cli-flags-dry).

```hcl
terragrunt = {
  terraform {
    # Force Terraform to not ask for input value if some variables are undefined.
    extra_arguments "disable_input" {
      commands  = ["${get_terraform_commands_that_need_input()}"]
      arguments = ["-input=false"]
    }
  }
}
```

#### get_terraform_commands_that_need_locking

`get_terraform_commands_that_need_locking()`

Returns the list of terraform commands that accept -lock-timeout parameter. This function is used when defining [extra_arguments](#keep-your-cli-flags-dry).

```hcl
terragrunt = {
  terraform {
    # Force Terraform to keep trying to acquire a lock for up to 20 minutes if someone else already has the lock
    extra_arguments "retry_lock" {
      commands  = ["${get_terraform_commands_that_need_locking()}"]
      arguments = ["-lock-timeout=20m"]
    }
  }
}
```

_Note: Functions that return a list of values must be used in a single declaration like:_

```hcl
commands = ["${get_terraform_commands_that_need_vars()}"]

# which result in:
commands = ["apply", "console", "destroy", "import", "plan", "push", "refresh"]

# We do not recommend using them in string composition like:
commands = "Some text ${get_terraform_commands_that_need_locking()}"

# which result in something useless like:
commands = "Some text [apply destroy import init plan refresh taint untaint]"
```


#### get_aws_account_id

`get_aws_account_id()` returns the AWS account id associated with the current set of credentials. Example:

```hcl
terragrunt = {
  remote_state {
    backend = "s3"
    config {
      bucket = "mycompany-${get_aws_account_id()}"
    }
  }
}
```

This allows uniqueness of the storage bucket per AWS account (since bucket name must be globally unique).

It is also possible to configure variables specifically based on the account used:

```
terragrunt = {
  terraform = {
    ...

    extra_arguments "common_var" {
      commands = ["${get_terraform_commands_that_need_vars()}"]
      arguments = ["-var-file=${get_aws_account_id()}.tfvars"]
    }
  }
}
```


### CLI Options

Terragrunt forwards all arguments and options to Terraform. The only exceptions are `--version` and arguments that
start with the prefix `--terragrunt-`. The currently available options are:

* `--terragrunt-config`: A custom path to the `terraform.tfvars` file. May also be specified via the `TERRAGRUNT_CONFIG`
  environment variable. The default path is `terraform.tfvars` in the current directory (see
  [Configuration](#configuration) for a slightly more nuanced explanation). This argument is not
  used with the `apply-all`, `destroy-all`, `output-all` and `plan-all` commands.

* `--terragrunt-tfpath`: A custom path to the Terraform binary. May also be specified via the `TERRAGRUNT_TFPATH`
  environment variable. The default is `terraform` in a directory on your PATH.

* `--terragrunt-non-interactive`: Don't show interactive user prompts. This will default the answer for all prompts to
  'yes'. Useful if you need to run Terragrunt in an automated setting (e.g. from a script).

* `--terragrunt-working-dir`: Set the directory where Terragrunt should execute the `terraform` command. Default is the
  current working directory. Note that for the `apply-all` and `destroy-all` directories, this parameter has a different
  meaning: Terragrunt will apply or destroy all the Terraform modules in the subfolders of the
  `terragrunt-working-dir`, running `terraform` in the root of each module it finds.

* `--terragrunt-source`: Download Terraform configurations from the specified source into a temporary folder, and run
  Terraform in that temporary folder. May also be specified via the `TERRAGRUNT_SOURCE` environment variable. The
  source should use the same syntax as the [Terraform module source](https://www.terraform.io/docs/modules/sources.html)
  parameter. This argument is not used with the `apply-all`, `destroy-all`, `output-all` and `plan-all` commands.

* `--terragrunt-source-update`: Delete the contents of the temporary folder before downloading Terraform source code
  into it.

* `--terragrunt-ignore-dependency-errors`: `*-all` commands continue processing components even if a dependency fails


### Configuration

Terragrunt configuration is defined in a `terraform.tfvars` file in a `terragrunt = { ... }` block.

For example:

```hcl
terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }

  dependencies {
    paths = ["../vpc", "../mysql", "../redis"]
  }
}
```

Terragrunt figures out the path to its config file according to the following rules:		
 		
 1. The value of the `--terragrunt-config` command-line option, if specified.
 1. The value of the `TERRAGRUNT_CONFIG` environment variable, if defined.
 1. A `terraform.tfvars` file in the current working directory, if it exists.
 1. If none of these are found, exit with an error.
 		
 The `--terragrunt-config` parameter is only used by Terragrunt and has no effect on which variable files are loaded
 by Terraform. Terraform will automatically read variables from a file named `terraform.tfvars`, but if you want it
 to read variables from some other .tfvars file, you must pass it in using the `--var-file` argument:		
 
 ```bash
  terragrunt plan --terragrunt-config example.tfvars --var-file example.tfvars		
 ```


##### Previous Versions of Terragrunt

Terragrunt v0.11.x and earlier defined the config in a .terragrunt file. Note that the .terragrunt format
is now deprecated. You will get a warning in your logs every time you run Terragrunt with a .terragrunt file,
and we will eventually stop supporting this older format.

### Migrating from Terragrunt v0.11.x and Terraform 0.8.x and older

After we released support for Terraform 0.9.x, we wrote a guide on 
[Upgrading to Terragrunt 0.12.x](_docs/migration_guides/upgrading_to_terragrunt_0.12.x.md).


### Developing terragrunt

#### Running locally

To run Terragrunt locally, use the `go run` command:

```bash
go run main.go plan
```

#### Dependencies

* Terragrunt uses `glide`, a vendor package management tool for golang. See the glide repo for
  [installation instructions](https://github.com/Masterminds/glide#install).

#### Running tests

**Note**: The tests in the `dynamodb` folder for Terragrunt run against a real AWS account and will add and remove
real data from DynamoDB. DO NOT hit `CTRL+C` while the tests are running, as this will prevent them from cleaning up
temporary tables and data in DynamoDB. We are not responsible for any charges you may incur.

Before running the tests, you must configure your [AWS credentials](#aws-credentials) and [AWS IAM policies](#aws-iam-policies).

To run all the tests:

```bash
go test -v -parallel 128 $(glide novendor)
```

To run only the tests in a specific package, such as the package `remote`:

```bash
cd remote
go test -v -parallel 128
```

And to run a specific test, such as `TestToTerraformRemoteConfigArgsNoBackendConfigs` in package `remote`:

```bash
cd remote
go test -v -parallel 128 -run TestToTerraformRemoteConfigArgsNoBackendConfigs
```


#### Debug logging

If you set the `TERRAGRUNT_DEBUG` environment variable to "true", the stack trace for any error will be printed to
stdout when you run the app.


#### Error handling

In this project, we try to ensure that:

1. Every error has a stacktrace. This makes debugging easier.
1. Every error generated by our own code (as opposed to errors from Go built-in functions or errors from 3rd party
   libraries) has a custom type. This makes error handling more precise, as we can decide to handle different types of
   errors differently.

To accomplish these two goals, we have created an `errors` package that has several helper methods, such as
`errors.WithStackTrace(err error)`, which wraps the given `error` in an Error object that contains a stacktrace. Under
the hood, the `errors` package is using the [go-errors](https://github.com/go-errors/errors) library, but this may
change in the future, so the rest of the code should not depend on `go-errors` directly.

Here is how the `errors` package should be used:

1. Any time you want to create your own error, create a custom type for it, and when instantiating that type, wrap it
   with a call to `errors.WithStackTrace`. That way, any time you call a method defined in the Terragrunt code, you
   know the error it returns already has a stacktrace and you don't have to wrap it yourself.
1. Any time you get back an error object from a function built into Go or a 3rd party library, immediately wrap it with
   `errors.WithStackTrace`. This gives us a stacktrace as close to the source as possible.
1. If you need to get back the underlying error, you can use the `errors.IsError` and `errors.Unwrap` functions.


#### Formatting

Every source file in this project should be formatted with `go fmt`. There are few helper scripts and targets in the
Makefile that can help with this (mostly taken from the [terraform repo](https://github.com/hashicorp/terraform/)):

1. `make fmtcheck`

   Checks to see if all source files are formatted. Exits 1 if there are unformatted files.
1. `make fmt`

    Formats all source files with `gofmt`.
1. `make install-pre-commit-hook`

    Installs a git pre-commit hook that will run all of the source files through `gofmt`.

To ensure that your changes get properly formatted, please install the git pre-commit hook with `make install-pre-commit-hook`.


#### Releasing new versions

To release a new version, just go to the [Releases Page](https://github.com/gruntwork-io/terragrunt/releases) and
create a new release. The CircleCI job for this repo has been configured to:

1. Automatically detect new tags.
1. Build binaries for every OS using that tag as a version number.
1. Upload the binaries to the release in GitHub.

See `circle.yml` and `_ci/build-and-push-release-asset.sh` for details.


### License

This code is released under the MIT License. See LICENSE.txt.
