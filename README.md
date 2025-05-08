# terraform-provider-utility

A simple Terraform provider to download files from a given URL with optional HTTP headers.

## Example Usage

```hcl
terraform {
  required_providers {
    utility = {
      source  = "frontiersgg/utility"
      version = "~> 0.1"
    }
  }
}

resource "utility_file_downloader" "example" {
  url      = "https://example.com/file.zip"
  filename = "${path.module}/file.zip"

  headers = {
    Authorization = "Bearer <token>"
  }
}
```

## Build

```shell
make build
```

## Development Overrides for Provider Developers

Local configuration file:

* Windows: `%APPDATA%\terraform.rc`
* Other OS: `~/.terraformrc`

```hcl
provider_installation {
  # Override providers.
  # This disables versioning and checksum validation of the provider
  # and forces Terraform to look for the provider in a given directory.
  dev_overrides {
    "registry.terraform.io/frontiersgg/utility" = "X:/path/to/terraform-provider-utility"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

> Provider file built on Windows must have an `.exe` extension.

## Reference

* https://github.com/hashicorp/terraform-provider-scaffolding-framework
* https://developer.hashicorp.com/terraform/cli/config/config-file
