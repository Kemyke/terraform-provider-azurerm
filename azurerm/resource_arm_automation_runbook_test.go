package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMAutomationRunbook_testScript(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureRMAutomationRunbook_testScript(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMAutomationRunbookDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMAutomationRunbookExists("azurerm_automation_runbook.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureRMAutomationRunbookDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).automationRunbookClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_automation_runbook" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		accName := rs.Primary.Attributes["account_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(resourceGroup, accName, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Automation Runbook still exists:\n%#v", resp)
		}
	}

	return nil
}

func testCheckAzureRMAutomationRunbookExists(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		accName := rs.Primary.Attributes["account_name"]

		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Automation Runbook: '%s'", name)
		}

		conn := testAccProvider.Meta().(*ArmClient).automationRunbookClient

		resp, err := conn.Get(resourceGroup, accName, name)

		if err != nil {
			return fmt.Errorf("Bad: Get on automationRunbookClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Automation Runbook '%s' (resource group: '%s') does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testAccAzureRMAutomationRunbook_testScript(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
 name = "acctestRG"
 location = "North Europe"
}

resource "azurerm_automation_account" "test" {
  name                = "acctest"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku {
        name = "Free"
  }
}

resource "azurerm_automation_runbook" "test" {
  name	 	      = "Get-AzureVMTutorial"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
 
  account_name        = "${azurerm_automation_account.test.name}"
  logVerbose	      = "true"
  logProgress	      = "true"
  description	      = "This is a test runbook for terraform acceptance test"
  runbookType	      = "PowerShellWorkflow"
  publishContentLink {
	uri = "https://raw.githubusercontent.com/Azure/azure-quickstart-templates/master/101-automation-runbook-getvms/Runbooks/Get-AzureVMTutorial.ps1"
  }
}
`)
}
