package openstack

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DeviaVir/terraform/helper/resource"
	"github.com/DeviaVir/terraform/helper/schema"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	nfloatingips "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

func resourceComputeFloatingIPAssociateV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceComputeFloatingIPAssociateV2Create,
		Read:   resourceComputeFloatingIPAssociateV2Read,
		Delete: resourceComputeFloatingIPAssociateV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"floating_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"wait_until_associated": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceComputeFloatingIPAssociateV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	computeClient, err := config.computeV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}

	floatingIP := d.Get("floating_ip").(string)
	fixedIP := d.Get("fixed_ip").(string)
	instanceId := d.Get("instance_id").(string)

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: floatingIP,
		FixedIP:    fixedIP,
	}
	log.Printf("[DEBUG] Associate Options: %#v", associateOpts)

	err = floatingips.AssociateInstance(computeClient, instanceId, associateOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error associating Floating IP: %s", err)
	}

	// This API call should be synchronous, but we've had reports where it isn't.
	// If the user opted in to wait for association, then poll here.
	var waitUntilAssociated bool
	if v, ok := d.GetOkExists("wait_until_associated"); ok {
		if wua, ok := v.(bool); ok {
			waitUntilAssociated = wua
		}
	}

	if waitUntilAssociated {
		log.Printf("[DEBUG] Waiting until %s is associated with %s", instanceId, floatingIP)

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"NOT_ASSOCIATED"},
			Target:     []string{"ASSOCIATED"},
			Refresh:    resourceComputeFloatingIPAssociateV2CheckAssociation(computeClient, instanceId, floatingIP),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      0,
			MinTimeout: 3 * time.Second,
		}

		_, err := stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	// There's an API call to get this information, but it has been
	// deprecated. The Neutron API could be used, but I'm trying not
	// to mix service APIs. Therefore, a faux ID will be used.
	id := fmt.Sprintf("%s/%s/%s", floatingIP, instanceId, fixedIP)
	d.SetId(id)

	return resourceComputeFloatingIPAssociateV2Read(d, meta)
}

func resourceComputeFloatingIPAssociateV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	computeClient, err := config.computeV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}

	// Obtain relevant info from parsing the ID
	floatingIP, instanceId, fixedIP, err := parseComputeFloatingIPAssociateId(d.Id())
	if err != nil {
		return err
	}

	// Now check and see whether the floating IP still exists.
	// First try to do this by querying the Network API.
	networkEnabled := true
	networkClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		networkEnabled = false
	}

	var exists bool
	if networkEnabled {
		log.Printf("[DEBUG] Checking for Floating IP existence via Network API")
		exists, err = resourceComputeFloatingIPAssociateV2NetworkExists(networkClient, floatingIP)
	} else {
		log.Printf("[DEBUG] Checking for Floating IP existence via Compute API")
		exists, err = resourceComputeFloatingIPAssociateV2ComputeExists(computeClient, floatingIP)
	}

	if err != nil {
		return err
	}

	if !exists {
		d.SetId("")
	}

	// Next, see if the instance still exists
	instance, err := servers.Get(computeClient, instanceId).Extract()
	if err != nil {
		if CheckDeleted(d, err, "instance") == nil {
			return nil
		}
	}

	// Finally, check and see if the floating ip is still associated with the instance.
	var associated bool
	for _, networkAddresses := range instance.Addresses {
		for _, element := range networkAddresses.([]interface{}) {
			address := element.(map[string]interface{})
			if address["OS-EXT-IPS:type"] == "floating" && address["addr"] == floatingIP {
				associated = true
			}
		}
	}

	if !associated {
		d.SetId("")
	}

	// Set the attributes pulled from the composed resource ID
	d.Set("floating_ip", floatingIP)
	d.Set("instance_id", instanceId)
	d.Set("fixed_ip", fixedIP)
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceComputeFloatingIPAssociateV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	computeClient, err := config.computeV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}

	floatingIP := d.Get("floating_ip").(string)
	instanceId := d.Get("instance_id").(string)

	disassociateOpts := floatingips.DisassociateOpts{
		FloatingIP: floatingIP,
	}
	log.Printf("[DEBUG] Disssociate Options: %#v", disassociateOpts)

	err = floatingips.DisassociateInstance(computeClient, instanceId, disassociateOpts).ExtractErr()
	if err != nil {
		return CheckDeleted(d, err, "floating ip association")
	}

	return nil
}

func parseComputeFloatingIPAssociateId(id string) (string, string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) < 3 {
		return "", "", "", fmt.Errorf("Unable to determine floating ip association ID")
	}

	floatingIP := idParts[0]
	instanceId := idParts[1]
	fixedIP := idParts[2]

	return floatingIP, instanceId, fixedIP, nil
}

func resourceComputeFloatingIPAssociateV2NetworkExists(networkClient *gophercloud.ServiceClient, floatingIP string) (bool, error) {
	listOpts := nfloatingips.ListOpts{
		FloatingIP: floatingIP,
	}
	allPages, err := nfloatingips.List(networkClient, listOpts).AllPages()
	if err != nil {
		return false, err
	}

	allFips, err := nfloatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return false, err
	}

	if len(allFips) > 1 {
		return false, fmt.Errorf("There was a problem retrieving the floating IP")
	}

	if len(allFips) == 0 {
		return false, nil
	}

	return true, nil
}

func resourceComputeFloatingIPAssociateV2ComputeExists(computeClient *gophercloud.ServiceClient, floatingIP string) (bool, error) {
	// If the Network API isn't available, fall back to the deprecated Compute API.
	allPages, err := floatingips.List(computeClient).AllPages()
	if err != nil {
		return false, err
	}

	allFips, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return false, err
	}

	for _, f := range allFips {
		if f.IP == floatingIP {
			return true, nil
		}
	}

	return false, nil
}

func resourceComputeFloatingIPAssociateV2CheckAssociation(
	computeClient *gophercloud.ServiceClient, instanceId, floatingIP string) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		instance, err := servers.Get(computeClient, instanceId).Extract()
		if err != nil {
			return instance, "", err
		}

		var associated bool
		for _, networkAddresses := range instance.Addresses {
			for _, element := range networkAddresses.([]interface{}) {
				address := element.(map[string]interface{})
				if address["OS-EXT-IPS:type"] == "floating" && address["addr"] == floatingIP {
					associated = true
				}
			}
		}

		if associated {
			return instance, "ASSOCIATED", nil
		}

		return instance, "NOT_ASSOCIATED", nil
	}
}
