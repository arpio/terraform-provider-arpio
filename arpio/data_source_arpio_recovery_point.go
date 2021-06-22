package arpio

import (
	"fmt"
	"github.com/arpio/arpio-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

func dataSourceArpioRecoveryPoint() *schema.Resource {
	//goland:noinspection GoDeprecation
	return &schema.Resource{
		Read:        dataSourceArpioRecoveryPointRead,
		Description: "Identifies an Arpio recovery point that can be used to recover stateful resources",
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Arpio application resource the recovery point was created for",
			},
			"timestamp": {
				Type:        schema.TypeString,
				Description: "Point in time to find the nearest existing recovery point (RFC 3339 format)",
				Optional:    true,
			},
			"timestamp_min": {
				Type:        schema.TypeString,
				Description: "Select only recovery points created on or after this point in time (RFC 3339 format)",
				Optional:    true,
			},
			"timeout": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     0,
				Description: "Duration to wait for a matching recovery point to exist",
			},
			"primary_account_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: false,
				Required: false,
			},
			"primary_region": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: false,
				Required: false,
			},
			"recovery_account_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: false,
				Required: false,
			},
			"recovery_region": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: false,
				Required: false,
			},
		},
	}
}

func dataSourceArpioRecoveryPointRead(d *schema.ResourceData, m interface{}) error {
	am := m.(*ProviderMetadata)

	if d.HasChange("app_id") || d.HasChange("timestamp") || d.HasChange("timestamp_min") {
		appID := d.Get("app_id")

		app, err := am.Client.GetApp(appID.(string))
		if err != nil {
			return err
		}

		syncPair := arpio.NewSyncPair(app.SourceAwsAccountID, app.SourceRegion, app.TargetAwsAccountID, app.TargetRegion)

		timeout, err := time.ParseDuration(d.Get("timeout").(string))
		if err != nil {
			return fmt.Errorf("error parsing timeout: %s", err)
		}

		timestampMin := d.Get("timestamp_min").(string)
		timestampMax := d.Get("timestamp").(string)

		tsMin, err := ParseRFC3339Timestamp(timestampMin)
		if err != nil {
			return err
		}
		tsMax, err := ParseRFC3339Timestamp(timestampMax)
		if err != nil {
			return err
		}

		rp, err := am.Client.MustFindLatestRecoveryPoint(syncPair, tsMin, tsMax, timeout)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Found recovery point %s at timestamp %s", rp.RecoveryPointID, rp.Timestamp)

		if err := d.Set("primary_account_id", app.SourceAwsAccountID); err != nil {
			return err
		}
		if err := d.Set("primary_region", app.SourceRegion); err != nil {
			return err
		}
		if err := d.Set("recovery_account_id", app.TargetAwsAccountID); err != nil {
			return err
		}
		if err := d.Set("recovery_region", app.TargetRegion); err != nil {
			return err
		}
		d.SetId(rp.RecoveryPointID)
	}

	return nil
}
