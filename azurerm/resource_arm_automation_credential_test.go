package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMAutomationCredential_testScript(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureRMAutomationCredential_testScript(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMAutomationCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMAutomationCredentialExists("azurerm_automation_credential.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureRMAutomationCredentialDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).automationCredentialClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_automation_credential" {
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
			return fmt.Errorf("Automation Credential still exists:\n%#v", resp)
		}
	}

	return nil
}

func testCheckAzureRMAutomationCredentialExists(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["credential_name"]
		accName := rs.Primary.Attributes["account_name"]

		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Automation Credential: '%s'", name)
		}

		conn := testAccProvider.Meta().(*ArmClient).automationCredentialClient

		resp, err := conn.Get(resourceGroup, accName, name)

		if err != nil {
			return fmt.Errorf("Bad: Get on automationCredentialClient: %s\nName: %s, Account name: %s, Resource group: %s OBJECT: %+v", err, name, accName, resourceGroup, rs.Primary)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Automation Credential '%s' (resource group: '%s') does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testAccAzureRMAutomationCredential_testScript(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
 name = "acctestRG"
 location = "North Europe"
}

resource "azurerm_automation_credential" "test" {
  credential_name     = "DefaultAzureCredential"
  resource_group_name = "${azurerm_resource_group.test.name}"
  account_name        = "${azurerm_automation_account.test.name}"
  user_name           = "kemy"
  password            = "pwd"
  description         = "This is a test credential for terraform acceptance test"
}

resource "azurerm_automation_account" "test" {
  name                = "acctest"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku {
        name = "Free"
  }
}

`)
}
