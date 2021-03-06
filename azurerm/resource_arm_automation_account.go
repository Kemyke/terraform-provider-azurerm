package azurerm

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/arm/automation"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceArmAutomationAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmAutomationAccountCreateUpdate,
		Read:   resourceArmAutomationAccountRead,
		Update: resourceArmAutomationAccountCreateUpdate,
		Delete: resourceArmAutomationAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: azureRMNormalizeLocation,
			},

			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sku": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
							ValidateFunc: validation.StringInSlice([]string{
								string(automation.Free),
								string(automation.Basic),
							}, true),
						},
					},
				},
				Set: resourceAzureRMAutomationAccountSkuHash,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAzureRMAutomationAccountSkuHash(v interface{}) int {
	m := v.(map[string]interface{})

	name := m["name"].(string)

	return hashcode.String(name)
}

func resourceArmAutomationAccountCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationAccountClient
	log.Printf("[INFO] preparing arguments for AzureRM Automation Account creation.")

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})

	sku := expandSku(d)

	parameters := automation.AccountCreateOrUpdateParameters{
		AccountCreateOrUpdateProperties: &automation.AccountCreateOrUpdateProperties{
			Sku: &sku,
		},

		Name:     &name,
		Location: &location,
		Tags:     expandTags(tags),
	}

	_, err := client.CreateOrUpdate(resGroup, name, parameters)
	if err != nil {
		return err
	}

	read, err := client.Get(resGroup, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Account '%s' (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmAutomationAccountRead(d, meta)
}

func resourceArmAutomationAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationAccountClient
	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["automationAccounts"]

	resp, err := client.Get(resGroup, name)
	if err != nil {
		if responseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AzureRM Automation Account '%s': %s", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("location", azureRMNormalizeLocation(*resp.Location))
	d.Set("resource_group_name", resGroup)
	flattenAndSetSku(d, resp.Sku)

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmAutomationAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationAccountClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["automationAccounts"]

	resp, err := client.Delete(resGroup, name)

	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Automation Account '%s': %+v", name, err)
	}

	return nil
}

func flattenAndSetSku(d *schema.ResourceData, sku *automation.Sku) {
	results := schema.Set{
		F: resourceAzureRMAutomationAccountSkuHash,
	}

	result := map[string]interface{}{}
	result["name"] = string(sku.Name)
	results.Add(result)

	d.Set("sku", &results)
}

func expandSku(d *schema.ResourceData) automation.Sku {
	inputs := d.Get("sku").(*schema.Set).List()
	input := inputs[0].(map[string]interface{})
	name := automation.SkuNameEnum(input["name"].(string))

	sku := automation.Sku{
		Name: name,
	}

	return sku
}
