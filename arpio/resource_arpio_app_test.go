package arpio

import (
	"fmt"
	"github.com/arpio/arpio-client-go"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccArpioAppProtect(t *testing.T) {
	appName := fmt.Sprintf("%s %d", t.Name(), time.Now().UnixNano())
	defer testAccCleanupApps(appName)

	//goland:noinspection ALL
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckArpioAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig(appName, "60"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAttrs("arpio_app.site", nil, appName, "60"),
				),
			},
			{
				Config: testAccAppConfig(appName, "120"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAttrs("arpio_app.site", nil, appName, "120"),
				),
			},
		},
	})
}

func TestAccArpioAppRecover(t *testing.T) {
	appName := fmt.Sprintf("%s %d", t.Name(), time.Now().UnixNano())
	defer testAccCleanupApps(appName)

	// Holds the ID of the app created before CRUD steps execute on tested
	// configs.  This simulates the app the user created in the primary
	// region, before attempting to recover in this test.
	var existingAppId StringHolder

	//goland:noinspection ALL
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckArpioAppDestroy,
		Steps: []resource.TestStep{
			{
				// During a recovery it's common that there isn't any local state
				// for the Arpio application.  Test that an existing application
				// is found by name and adopted into the state when it exists.
				PreConfig: testAccCreateApp(t, appName, "120", &existingAppId),
				Config:    testAccAppConfig(appName, "60"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAttrs("arpio_app.site", &existingAppId, appName, "60"),
				),
			},
		},
	})
}

func TestAccArpioAppRecoverMultipleExist(t *testing.T) {
	appName := fmt.Sprintf("%s %d", t.Name(), time.Now().UnixNano())
	defer testAccCleanupApps(appName)

	//goland:noinspection ALL
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckArpioAppDestroy,
		Steps: []resource.TestStep{
			{
				// If multiple apps exist with the recovery app's name, create
				// returns an error and ID is not set in the state.
				PreConfig: func() {
					testAccCreateApp(t, appName, "120", nil)()
					testAccCreateApp(t, appName, "180", nil)()
				},
				Config:      testAccAppConfig(appName, "60"),
				ExpectError: regexp.MustCompile("more than one Arpio app already exists with the name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppNotExists("arpio_app.site"),
				),
			},
		},
	})
}

func testAccAppConfig(appName, rpo string) string {
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
		
		resource "arpio_app" "site" {
			name                = "%s"
			rpo                 = %s
			primary_account_id  = "%s"
			primary_region      = "%s"
			recovery_account_id = "%s"
			recovery_region     = "%s"
			notification_emails = ["admin@example.com", "other@example.com"]
			resources {
				arns = [
					"arn:aws:ec2:%s:%s:instance/i-abcdefg1234567890",
					"arn:aws:ec2:%s:%s:instance/i-bcdefgh1234567890",
				]
				tags = {
					"ArpioProtectThis" = "true"
					"ArpioAnyValue"    = ""
				}
			}
		}
		`,
		accountID, apiKeyID, apiKeySecret, apiURL,
		appName, rpo,
		sourceAwsAccountID, sourceRegion, targetAwsAccountID, targetRegion,
		sourceRegion, sourceAwsAccountID,
		sourceRegion, sourceAwsAccountID,
	)
}

func testAccCreateApp(t *testing.T, name, rpo string, createdAppID *StringHolder) func() {
	return func() {
		ac := testAccProvider.Meta().(*ProviderMetadata).Client

		accountID := os.Getenv(ArpioAccountIDEnv)
		sourceAwsAccountID := os.Getenv(ArpioTestSourceAwsAccountIDEnv)
		sourceRegion := os.Getenv(ArpioTestSourceApiRegionEnv)
		targetAwsAccountID := os.Getenv(ArpioTestTargetAwsAccountIDEnv)
		targetRegion := os.Getenv(ArpioTestTargetApiRegionEnv)

		a := ac.NewApp()
		a.Name = name
		a.AccountID = accountID
		a.SourceAwsAccountID = sourceAwsAccountID
		a.SourceRegion = sourceRegion
		a.TargetAwsAccountID = targetAwsAccountID
		a.TargetRegion = targetRegion
		a.NotificationEmails = []string{}
		a.SelectionRules = []arpio.SelectionRule{}

		rpoInt, err := strconv.Atoi(rpo)
		if err != nil {
			t.Fatalf("Error parsing RPO %s: %s", rpo, err)
		}
		a.RPO = rpoInt * 60

		app, err := ac.CreateApp(a)
		if err != nil {
			t.Fatalf("Error creating app: %s", err)
		}

		if createdAppID != nil {
			createdAppID.S = app.AppID
		}
	}
}

func testAccCheckAppNotExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return nil
		}

		if rs.Primary.ID != "" {
			return fmt.Errorf("ID set")
		}

		return nil
	}
}

func testAccCheckAppAttrs(n string, appID *StringHolder, name, rpo string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set")
		}

		if appID != nil && appID.S != "" {
			if rs.Primary.ID != appID.S {
				return fmt.Errorf("ID %s != %s", appID.S, rs.Primary.ID)
			}
		}

		// Expected values
		sourceAwsAccountID := os.Getenv(ArpioTestSourceAwsAccountIDEnv)
		sourceRegion := os.Getenv(ArpioTestSourceApiRegionEnv)
		targetAwsAccountID := os.Getenv(ArpioTestTargetAwsAccountIDEnv)
		targetRegion := os.Getenv(ArpioTestTargetApiRegionEnv)

		if err := checkAttrString(rs, "name", name); err != nil {
			return err
		}
		if err := checkAttrString(rs, "rpo", rpo); err != nil {
			return err
		}
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
		if err := checkAttrString(rs, "resources.0.arns.#", "2"); err != nil {
			return fmt.Errorf("expected 2 ARN rules")
		}
		if err := checkAttrString(rs, "resources.0.tags.%", "2"); err != nil {
			return fmt.Errorf("expected 2 tag rules")
		}

		return nil
	}
}

func testAccCheckArpioAppDestroy(s *terraform.State) error {
	ac := testAccProvider.Meta().(*ProviderMetadata).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arpio_app" {
			continue
		}

		// Try to find the resource
		app, err := ac.GetApp(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error getting app: %s", rs.Primary.ID)
		}
		if app != nil {
			return fmt.Errorf("App should not exist: %s", rs.Primary.ID)
		}
	}

	return nil
}
