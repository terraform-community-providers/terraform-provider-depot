---
page_title: "Depot Provider"
---

# Depot Provider

This provider is used to interact with the many resources supported by [Depot](https://depot.dev).

## Authentication

This provider requires a Depot API token in order to manage resources.

To manage the full selection of resources, provide a user token from an account with appropriate permissions.

There are several ways to provide the required token:

* **Set the `token` argument in the provider configuration**. You can set the `token` argument in the provider configuration. Use an input variable for the token.
* **Set the `DEPOT_TOKEN` environment variable**. The provider can read the `DEPOT_TOKEN` environment variable and the token stored there to authenticate.

## Example Usage

```terraform
provider "depot" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `token` (String) The token used to authenticate with Depot.
