package arpio

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccArpioRecoveryPoint(t *testing.T) {
	appName := fmt.Sprintf("%s %d", t.Name(), time.Now().UnixNano())
	defer testAccCleanupApps(appName)

	//goland:noinspection ALL
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckArpioAppDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRecoveryPointConfig(appName, "", "2000-01-01T00:00:00Z", "0s"),
				ExpectError: regexp.MustCompile("no recovery points on or before"),
			},
			{
				Config:      testAccRecoveryPointConfig(appName, "", "2000-01-01T00:00:00Z", "1ms"),
				ExpectError: regexp.MustCompile("no recovery points on or before"),
			},
			{
				Config:      testAccRecoveryPointConfig(appName, "2100-01-01T00:00:00Z", "", "0s"),
				ExpectError: regexp.MustCompile("no recovery points on or after"),
			},
			{
				Config:      testAccRecoveryPointConfig(appName, "2100-01-01T00:00:00Z", "", "1ms"),
				ExpectError: regexp.MustCompile("no recovery points on or after"),
			},
			{
				Config:      testAccRecoveryPointConfig(appName, "2100-01-01T00:00:00Z", "2200-01-01T00:00:00Z", "0s"),
				ExpectError: regexp.MustCompile("no recovery points between"),
			},
			{
				Config:      testAccRecoveryPointConfig(appName, "2100-01-01T00:00:00Z", "2200-01-01T00:00:00Z", "1ms"),
				ExpectError: regexp.MustCompile("no recovery points between"),
			},
			{
				Config: testAccRecoveryPointConfig(appName, "2000-01-01T00:00:00Z", "", "5m"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.latest"),
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.time_constrained"),
				),
			},
			{
				Config: testAccRecoveryPointConfig(appName, "", "2100-01-01T00:00:00Z", "5m"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.latest"),
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.time_constrained"),
				),
			},
			{
				Config: testAccRecoveryPointConfig(appName, "2000-01-01T00:00:00Z", "2100-01-01T00:00:00Z", "5m"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.latest"),
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.time_constrained"),
				),
			},
			{
				Config: testAccRecoveryPointConfig(appName, "", "", "5m"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.latest"),
					testAccCheckRecoveryPointAttrs("data.arpio_recovery_point.time_constrained"),
				),
			},
		},
	})
}

func testAccRecoveryPointConfig(appName string, timestampMin, timestamp string, timeout string) string {
	accountID := os.Getenv(ArpioAccountIDEnv)
	apiKeyID := os.Getenv(ArpioApiKeyIDEnv)
	apiKeySecret := os.Getenv(ArpioApiKeySecretEnv)
	apiURL := os.Getenv(ArpioApiURLEnv)
	sourceAwsAccountID := os.Getenv(ArpioTestSourceAwsAccountIDEnv)
	sourceRegion := os.Getenv(ArpioTestSourceApiRegionEnv)
	targetAwsAccountID := os.Getenv(ArpioTestTargetAwsAccountIDEnv)
	targetRegion := os.Getenv(ArpioTestTargetApiRegionEnv)

	return fmt.Sprintf(`
		provider "arpio" {
			account_id     = "%s"
			api_key_id     = "%s"
			api_key_secret = "%s"
			api_url        = "%s"
		}
		
		resource "arpio_app" "empty" {
			name                = "%s"
			rpo                 = 60
			primary_account_id  = "%s"
			primary_region      = "%s"
			recovery_account_id = "%s"
			recovery_region     = "%s"
		}

		data "arpio_recovery_point" "latest" {
			app_id  = arpio_app.empty.id
			timeout = "%s"
		}

		data "arpio_recovery_point" "time_constrained" {
			app_id        = arpio_app.empty.id
			timestamp_min = "%s" 
			timestamp     = "%s" 
			timeout       = "%s"
		}
		`,
		accountID, apiKeyID, apiKeySecret, apiURL,
		appName,
		sourceAwsAccountID, sourceRegion, targetAwsAccountID, targetRegion,
		timeout,
		timestampMin, timestamp, timeout,
	)
}

func testAccCheckRecoveryPointAttrs(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set")
		}

		if rs.Primary.Attributes["app_id"] == "" {
			return fmt.Errorf("No app_id set")
		}

		// Expected values
		sourceAwsAccountID := os.Getenv(ArpioTestSourceAwsAccountIDEnv)
		sourceRegion := os.Getenv(ArpioTestSourceApiRegionEnv)
		targetAwsAccountID := os.Getenv(ArpioTestTargetAwsAccountIDEnv)
		targetRegion := os.Getenv(ArpioTestTargetApiRegionEnv)

		if err := checkAttrString(rs, "primary_account_id", sourceAwsAccountID); err != nil {
			return err
		}
		if err := checkAttrString(rs, "primary_region", sourceRegion); err != nil {
			return err
		}
		if err := checkAttrString(rs, "recovery_account_id", targetAwsAccountID); err != nil {
			return err
		}
		if err := checkAttrString(rs, "recovery_region", targetRegion); err != nil {
			return err
		}

		return nil
	}
}
