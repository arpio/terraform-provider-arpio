module github.com/arpio/terraform-provider-arpio

go 1.13

require (
	github.com/arpio/arpio-client-go v0.1.3
	github.com/arpio/patchenv v1.0.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
)

// Uncomment for local development.
//replace github.com/arpio/arpio-client-go => ../arpio-client-go
