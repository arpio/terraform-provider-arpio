package arpio

import (
	ac "github.com/arpio/arpio-client-go"
	"github.com/arpio/patchenv"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

type ProviderMetadata struct {
	Client *ac.Client
}

func Provider() *schema.Provider {
	//goland:noinspection ALL
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of the Arpio API",
				DefaultFunc: schema.EnvDefaultFunc(ArpioApiURLEnv, ac.ArpioURL),
			},
			"api_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "ID of the Arpio API key authorized to access the specified Arpio account",
				DefaultFunc: schema.EnvDefaultFunc(ArpioApiKeyIDEnv, nil),
			},
			"api_key_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Secret for the Arpio API key authorized to access the specified Arpio account",
				DefaultFunc: schema.EnvDefaultFunc(ArpioApiKeySecretEnv, nil),
			},
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Arpio account that protects and recovers stateful resources",
				DefaultFunc: schema.EnvDefaultFunc(ArpioAccountIDEnv, nil),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"arpio_recovery_point": dataSourceArpioRecoveryPoint(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"arpio_app": resourceArpioApp(),
		},
		ConfigureFunc: arpioConfigure,
	}
}

func arpioConfigure(d *schema.ResourceData) (m interface{}, err error) {
	if err := patchenv.Patch(); err != nil {
		log.Printf("[ERROR] %s", err)
	}
	if err := WaitForDebuggerAttach(); err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}

	// Configure the Arpio client
	apiURL := d.Get("api_url").(string)
	apiKeyID := d.Get("api_key_id").(string)
	apiKeySecret := d.Get("api_key_secret").(string)
	accountID := d.Get("account_id").(string)

	client, err := ac.NewClient(apiURL, apiKeyID, apiKeySecret, accountID)
	if err != nil {
		return nil, err
	}

	metadata := &ProviderMetadata{
		Client: client,
	}
	return metadata, nil
}
