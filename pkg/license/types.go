package license

import "time"

type AzureInstanceMetadata struct {
	Compute Compute `json:"compute"`
	Network Network `json:"network"`
}
type OsProfile struct {
	AdminUsername                 string `json:"adminUsername"`
	ComputerName                  string `json:"computerName"`
	DisablePasswordAuthentication string `json:"disablePasswordAuthentication"`
}
type Plan struct {
	Name      string `json:"name"`
	Product   string `json:"product"`
	Publisher string `json:"publisher"`
}
type PublicKeys struct {
	KeyData string `json:"keyData"`
	Path    string `json:"path"`
}
type SecurityProfile struct {
	SecureBootEnabled string `json:"secureBootEnabled"`
	VirtualTpmEnabled string `json:"virtualTpmEnabled"`
}
type ImageReference struct {
	ID        string `json:"id"`
	Offer     string `json:"offer"`
	Publisher string `json:"publisher"`
	Sku       string `json:"sku"`
	Version   string `json:"version"`
}
type DiffDiskSettings struct {
	Option string `json:"option"`
}
type EncryptionSettings struct {
	Enabled string `json:"enabled"`
}
type Image struct {
	URI string `json:"uri"`
}
type ManagedDisk struct {
	ID                 string `json:"id"`
	StorageAccountType string `json:"storageAccountType"`
}
type Vhd struct {
	URI string `json:"uri"`
}
type OsDisk struct {
	Caching                 string             `json:"caching"`
	CreateOption            string             `json:"createOption"`
	DiffDiskSettings        DiffDiskSettings   `json:"diffDiskSettings"`
	DiskSizeGB              string             `json:"diskSizeGB"`
	EncryptionSettings      EncryptionSettings `json:"encryptionSettings"`
	Image                   Image              `json:"image"`
	ManagedDisk             ManagedDisk        `json:"managedDisk"`
	Name                    string             `json:"name"`
	OsType                  string             `json:"osType"`
	Vhd                     Vhd                `json:"vhd"`
	WriteAcceleratorEnabled string             `json:"writeAcceleratorEnabled"`
}
type ResourceDisk struct {
	Size string `json:"size"`
}
type StorageProfile struct {
	DataDisks      []any          `json:"dataDisks"`
	ImageReference ImageReference `json:"imageReference"`
	OsDisk         OsDisk         `json:"osDisk"`
	ResourceDisk   ResourceDisk   `json:"resourceDisk"`
}
type Compute struct {
	AzEnvironment              string          `json:"azEnvironment"`
	CustomData                 string          `json:"customData"`
	EvictionPolicy             string          `json:"evictionPolicy"`
	IsHostCompatibilityLayerVM string          `json:"isHostCompatibilityLayerVm"`
	LicenseType                string          `json:"licenseType"`
	Location                   string          `json:"location"`
	Name                       string          `json:"name"`
	Offer                      string          `json:"offer"`
	OsProfile                  OsProfile       `json:"osProfile"`
	OsType                     string          `json:"osType"`
	PlacementGroupID           string          `json:"placementGroupId"`
	Plan                       Plan            `json:"plan"`
	PlatformFaultDomain        string          `json:"platformFaultDomain"`
	PlatformUpdateDomain       string          `json:"platformUpdateDomain"`
	Priority                   string          `json:"priority"`
	Provider                   string          `json:"provider"`
	PublicKeys                 []PublicKeys    `json:"publicKeys"`
	Publisher                  string          `json:"publisher"`
	ResourceGroupName          string          `json:"resourceGroupName"`
	ResourceID                 string          `json:"resourceId"`
	SecurityProfile            SecurityProfile `json:"securityProfile"`
	Sku                        string          `json:"sku"`
	StorageProfile             StorageProfile  `json:"storageProfile"`
	SubscriptionID             string          `json:"subscriptionId"`
	Tags                       string          `json:"tags"`
	TagsList                   []any           `json:"tagsList"`
	UserData                   string          `json:"userData"`
	Version                    string          `json:"version"`
	VMID                       string          `json:"vmId"`
	VMScaleSetName             string          `json:"vmScaleSetName"`
	VMSize                     string          `json:"vmSize"`
	Zone                       string          `json:"zone"`
}
type IPAddress struct {
	PrivateIPAddress string `json:"privateIpAddress"`
	PublicIPAddress  string `json:"publicIpAddress"`
}
type Subnet struct {
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
}
type Ipv4 struct {
	IPAddress []IPAddress `json:"ipAddress"`
	Subnet    []Subnet    `json:"subnet"`
}
type Ipv6 struct {
	IPAddress []any `json:"ipAddress"`
}
type Interface struct {
	Ipv4       Ipv4   `json:"ipv4"`
	Ipv6       Ipv6   `json:"ipv6"`
	MacAddress string `json:"macAddress"`
}
type Network struct {
	Interface []Interface `json:"interface"`
}

type InstanceIdentityDocument struct {
	AccountID               string    `json:"accountId"`
	Architecture            string    `json:"architecture"`
	AvailabilityZone        string    `json:"availabilityZone"`
	BillingProducts         any       `json:"billingProducts"`
	DevpayProductCodes      any       `json:"devpayProductCodes"`
	MarketplaceProductCodes []string  `json:"marketplaceProductCodes"`
	ImageID                 string    `json:"imageId"`
	InstanceID              string    `json:"instanceId"`
	InstanceType            string    `json:"instanceType"`
	KernelID                any       `json:"kernelId"`
	PendingTime             time.Time `json:"pendingTime"`
	PrivateIP               string    `json:"privateIp"`
	RamdiskID               any       `json:"ramdiskId"`
	Region                  string    `json:"region"`
	Version                 string    `json:"version"`
}

type License struct {
	Users int `json:"users"`
}
