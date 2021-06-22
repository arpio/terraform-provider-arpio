package arpio

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider
var testAccProviderConfigure sync.Once

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"arpio": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatal(err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	vars := []string{
		ArpioApiURLEnv,
		ArpioApiKeyIDEnv,
		ArpioApiKeySecretEnv,
		ArpioAccountIDEnv,
		ArpioTestSourceAwsAccountIDEnv,
		ArpioTestSourceApiRegionEnv,
		ArpioTestTargetAwsAccountIDEnv,
		ArpioTestTargetApiRegionEnv,
	}
	for _, v := range vars {
		if val := os.Getenv(v); val == "" {
			t.Fatalf("%s must be set for acceptance tests", v)
		}
	}

	testAccProviderConfigure.Do(func() {
		// Configure the provider so it's usable in test functions that run
		// outside the normal Terraform CRUD context.
		err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func testAccCleanupApps(name string) {
	// We may fail in Configure, before the metadata gets set
	meta := testAccProvider.Meta()
	if meta == nil {
		return
	}

	ac := meta.(*ProviderMetadata).Client

	apps, err := ac.ListApps()
	if err == nil {
		for _, app := range apps {
			if app.Name == name {
				err := ac.DeleteApp(app.AppID)
				if err != nil {
					log.Printf("[WARN] error deleting app %q: %s", app.AppID, err)
				}
			}
		}
	}
}

func checkAttrString(rs *terraform.ResourceState, attrName, expectedValue string) error {
	v := rs.Primary.Attributes[attrName]
	if v != expectedValue {
		return fmt.Errorf("%s: %s != %s", attrName, v, expectedValue)
	}
	return nil
}
