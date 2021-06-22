package arpio

import (
	"fmt"
	ac "github.com/arpio/arpio-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sort"
)

func resourceArpioApp() *schema.Resource {
	//goland:noinspection GoDeprecation
	return &schema.Resource{
		Create: resourceArpioAppCreate,
		Read:   resourceArpioAppRead,
		Update: resourceArpioAppUpdate,
		Delete: resourceArpioAppDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rpo": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"primary_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"primary_region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"recovery_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"recovery_region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"notification_emails": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Email address of Arpio users who wish to receive notification emails",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resources": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Specifies rules for matching resources to protect",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arns": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "ARNs of stateful resources that Arpio should protect",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tags": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Tags matching stateful resources that Arpio should protect",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceArpioAppCreate(d *schema.ResourceData, m interface{}) error {
	am := m.(*ProviderMetadata)

	app := am.Client.NewApp()
	if err := setAppFromResourceData(d, &app); err != nil {
		return err
	}

	// If an application with the same name already exists, we'll update it
	// to match the resource's config and use it.  If there are multiple apps
	// with the same name, we'll fail the create and let the user sort it out.
	apps, err := am.Client.ListApps()
	if err != nil {
		return err
	}

	var existingAppID string
	for _, existingApp := range apps {
		if existingApp.Name == app.Name {
			if existingAppID != "" {
				return fmt.Errorf("more than one Arpio app already "+
					"exists with the name %q; use the Arpio web interface to "+
					"rename the unrelated apps, then retry the creation",
					app.Name)
			}
			existingAppID = existingApp.AppID
		}
	}

	if existingAppID != "" {
		// Update the existing app
		d.SetId(existingAppID)
		return resourceArpioAppUpdate(d, m)
	} else {
		// Create a new app
		createdApp, err := am.Client.CreateApp(app)
		if err != nil {
			return err
		}
		d.SetId(createdApp.AppID)
		return setResourceDataFromApp(d, app)
	}
}

func resourceArpioAppRead(d *schema.ResourceData, m interface{}) error {
	am := m.(*ProviderMetadata)

	app, err := am.Client.GetApp(d.Id())
	if err != nil {
		return err
	}

	if app.AppID == "" {
		d.SetId("")
		return nil
	}

	return setResourceDataFromApp(d, *app)
}

func resourceArpioAppUpdate(d *schema.ResourceData, m interface{}) error {
	am := m.(*ProviderMetadata)

	app, err := am.Client.GetApp(d.Id())
	if err != nil {
		return err
	}

	err = setAppFromResourceData(d, app)
	if err != nil {
		return err
	}

	updated, err := am.Client.UpdateApp(*app)
	if err != nil {
		return err
	}

	return setResourceDataFromApp(d, updated)
}

func resourceArpioAppDelete(d *schema.ResourceData, m interface{}) error {
	am := m.(*ProviderMetadata)
	return am.Client.DeleteApp(d.Id())
}

func setAppFromResourceData(d *schema.ResourceData, app *ac.App) error {
	emails := TypifyStringList(d.Get("notification_emails").([]interface{}))
	sort.Strings(emails)

	app.Name = d.Get("name").(string)
	app.RPO = d.Get("rpo").(int) * 60
	app.SourceAwsAccountID = d.Get("primary_account_id").(string)
	app.SourceRegion = d.Get("primary_region").(string)
	app.TargetAwsAccountID = d.Get("recovery_account_id").(string)
	app.TargetRegion = d.Get("recovery_region").(string)
	app.NotificationEmails = emails
	if err := setAppSelectionRulesFromResourceData(d, app); err != nil {
		return err
	}
	return nil
}

func setAppSelectionRulesFromResourceData(d *schema.ResourceData, app *ac.App) error {
	var rules []ac.SelectionRule

	if attrs, ok := GetFirstElementAsMap(d.Get("resources")); ok {
		// Include one ARN rule that covers all the specified ARNs
		arns := FromSetToStringList(attrs["arns"])
		if len(arns) > 0 {
			for _, arn := range arns {
				if !IsARN(arn) {
					return fmt.Errorf("%q is not a valid ARN ", arn)
				}
			}
			sort.Strings(arns)
			rules = append(rules, ac.NewArnRule(arns))
		}

		// Include one tag rule for each key and value
		tags := FromSetToStringMap(attrs["tags"])
		if len(tags) > 0 {
			for k, v := range tags {
				rules = append(rules, ac.NewTagRule(k, v))
			}
		}
	}

	app.SelectionRules = rules
	return nil
}

func setResourceDataFromApp(d *schema.ResourceData, app ac.App) error {
	syncPair := app.SyncPair()
	if err := d.Set("name", app.Name); err != nil {
		return err
	}
	if err := d.Set("rpo", app.RPO/60); err != nil {
		return err
	}
	if err := d.Set("primary_account_id", syncPair.Source.AccountID); err != nil {
		return err
	}
	if err := d.Set("primary_region", syncPair.Source.Region); err != nil {
		return err
	}
	if err := d.Set("recovery_account_id", syncPair.Target.AccountID); err != nil {
		return err
	}
	if err := d.Set("recovery_region", syncPair.Target.Region); err != nil {
		return err
	}
	if err := d.Set("notification_emails", app.NotificationEmails); err != nil {
		return err
	}
	if err := setResourceDataSelectionRulesFromApp(d, app); err != nil {
		return err
	}
	return nil
}

func setResourceDataSelectionRulesFromApp(d *schema.ResourceData, app ac.App) error {
	// Combine all the ARN rules into one list, all the tag rules into one map
	var arns []string
	tags := map[string]string{}

	for _, rule := range app.SelectionRules {
		switch rule.GetRuleType() {
		case ac.ArnRuleType:
			arnRule := rule.(ac.ArnRule)
			for _, arn := range arnRule.Arns {
				arns = append(arns, arn)
			}
		case ac.TagRuleType:
			tagRule := rule.(ac.TagRule)
			tags[tagRule.Name] = tagRule.Value
		}
	}

	sort.Strings(arns)

	if attrs, ok := GetFirstElementAsMap(d.Get("resources")); ok {
		attrs["arns"] = schema.NewSet(schema.HashString, UntypifyStringList(arns))
		attrs["tags"] = tags
		err := d.Set("resources", []interface{}{attrs})
		if err != nil {
			return err
		}
	}

	return nil
}
