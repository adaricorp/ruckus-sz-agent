package main

type ApConfig struct {
	BladeID           string `json:",omitempty"`
	GpsInfo           string `json:",omitempty"`
	ApID              string `json:",omitempty"`
	IP                string `json:",omitempty"`
	AltitudeUnit      string `json:",omitempty"`
	Description       string `json:",omitempty"`
	ApGroupID         string `json:",omitempty"`
	DeviceName        string `json:",omitempty"`
	ExtIP             string `json:",omitempty"`
	RogueEnabled      bool   `json:",omitempty"`
	UpTime            int    `json:",omitempty"`
	LastSeen          int64  `json:",omitempty"`
	Serial            string `json:",omitempty"`
	ConnectionStatus  string `json:",omitempty"`
	ZoneID            string `json:",omitempty"`
	Model             string `json:",omitempty"`
	Location          string `json:",omitempty"`
	AltitudeValue     int    `json:",omitempty"`
	FwVersion         string `json:",omitempty"`
	RegistrationState string `json:",omitempty"`
}

type ControlBladeConfig struct {
	DNSV6List               []interface{} `json:",omitempty"`
	PrimaryDNSServer        string        `json:",omitempty"`
	MemoryInfo              string        `json:",omitempty"`
	NumberOfPorts           string        `json:",omitempty"`
	AssociateAPs            int           `json:",omitempty"`
	ExportBindingInterfaces struct {
		Control    string `json:",omitempty"`
		Cluster    string `json:",omitempty"`
		Management string `json:",omitempty"`
	} `json:",omitempty"`
	Model             string      `json:",omitempty"`
	NumberOfProcesors string      `json:",omitempty"`
	ClientCount       int         `json:",omitempty"`
	ClusterEtherIP    string      `json:",omitempty"`
	BladeStateEnum    string      `json:",omitempty"`
	ApBindingEther    string      `json:",omitempty"`
	ApCount           int         `json:",omitempty"`
	RemovalStatus     interface{} `json:",omitempty"`
	IP                string      `json:",omitempty"`
	V6ClusterEtherIP  interface{} `json:",omitempty"`
	BindingIps        struct{}    `json:",omitempty"`
	UptimeSec         int         `json:",omitempty"`
	HotspotEnabled    bool        `json:",omitempty"`
	ExportBondTopo    struct{}    `json:",omitempty"`
	NetworkInterfaces []struct {
		Iface              string      `json:",omitempty"`
		InterfaceType      string      `json:",omitempty"`
		IPMode             string      `json:",omitempty"`
		PrimaryDNSServer   interface{} `json:",omitempty"`
		IP                 string      `json:",omitempty"`
		SubnetMask         string      `json:",omitempty"`
		SecondaryDNSServer interface{} `json:",omitempty"`
		DefaultGateway     bool        `json:",omitempty"`
		Gateway            string      `json:",omitempty"`
	} `json:",omitempty"`
	LastSeen                       interface{}   `json:",omitempty"`
	UserInterfaces                 interface{}   `json:",omitempty"`
	V6PrimaryDNSServer             string        `json:",omitempty"`
	UserNetworkInterfaces          []interface{} `json:",omitempty"`
	DisplaySystemCapacity          string        `json:",omitempty"`
	V6BindingIps                   struct{}      `json:",omitempty"`
	BridgeAll                      bool          `json:",omitempty"`
	InterfaceV6AutoRaIPList        interface{}   `json:",omitempty"`
	ExportBrTopo                   struct{}      `json:",omitempty"`
	HostName                       string        `json:",omitempty"`
	ResourcePlan                   string        `json:",omitempty"`
	Role                           string        `json:",omitempty"`
	V6DefaultGatewayInterface      string        `json:",omitempty"`
	V6WebEtherIP                   interface{}   `json:",omitempty"`
	V6ApEtherIP                    interface{}   `json:",omitempty"`
	ClusterBindingEther            string        `json:",omitempty"`
	DiskInfo                       string        `json:",omitempty"`
	Mac                            string        `json:",omitempty"`
	V6Interfaces                   interface{}   `json:",omitempty"`
	ModelOfProcesor                string        `json:",omitempty"`
	Vendor                         string        `json:",omitempty"`
	NumOfDataBlades                int           `json:",omitempty"`
	PortGroup                      string        `json:",omitempty"`
	StartTime                      int64         `json:",omitempty"`
	CertImportResult               string        `json:",omitempty"`
	Firmware                       string        `json:",omitempty"`
	DefaultGatewayInterface        string        `json:",omitempty"`
	Key                            string        `json:",omitempty"`
	DisplayModelName               string        `json:",omitempty"`
	HealthEnum                     string        `json:",omitempty"`
	V6StaticRoutes                 []interface{} `json:",omitempty"`
	DNSV4List                      []string      `json:",omitempty"`
	WebBindingEther                string        `json:",omitempty"`
	CreateTime                     int           `json:",omitempty"`
	DataBladeInfo                  interface{}   `json:",omitempty"`
	BondInterfaceDisplayNameMap    struct{}      `json:",omitempty"`
	ApNumbers                      int           `json:",omitempty"`
	BondTopo                       struct{}      `json:",omitempty"`
	UdiConstainsInvalidVlanInRange bool          `json:",omitempty"`
	DNSList                        []string      `json:",omitempty"`
	RegistrationState              string        `json:",omitempty"`
	RouteTrafficValid              bool          `json:",omitempty"`
	Statistics                     interface{}   `json:",omitempty"`
	ControlNatIP                   string        `json:",omitempty"`
	SubNetworkInterfaces           []interface{} `json:",omitempty"`
	BrTopo                         struct{}      `json:",omitempty"`
	CpVersion                      string        `json:",omitempty"`
	SecondaryDNSServer             string        `json:",omitempty"`
	RouteSeparationEnabled         bool          `json:",omitempty"`
	Routes                         interface{}   `json:",omitempty"`
	BaseMac                        string        `json:",omitempty"`
	SubNetworkInterfaceNameMap     struct{}      `json:",omitempty"`
	SubInterfaces                  interface{}   `json:",omitempty"`
	SerialNumber                   string        `json:",omitempty"`
	Fqdn                           string        `json:",omitempty"`
	InternalSubnetPrefix           string        `json:",omitempty"`
	V6NetworkInterfaces            []interface{} `json:",omitempty"`
	UserNetworkInterfaceNameMap    struct{}      `json:",omitempty"`
	StaticRoutes                   []interface{} `json:",omitempty"`
	External                       bool          `json:",omitempty"`
	Name                           string        `json:",omitempty"`
	ControlIPForCpComm             string        `json:",omitempty"`
	LicenseUtilization             int           `json:",omitempty"`
	NumberOfBonds                  interface{}   `json:",omitempty"`
	ApEtherIP                      string        `json:",omitempty"`
	Description                    string        `json:",omitempty"`
	InterfaceV6AutoDhcpIPList      interface{}   `json:",omitempty"`
	InterfaceV6IPList              interface{}   `json:",omitempty"`
	BindingMacs                    struct{}      `json:",omitempty"`
	UserInterfaceIPv6ForCP         string        `json:",omitempty"`
	BindingInterfaces              struct{}      `json:",omitempty"`
	V6SecondaryDNSServer           string        `json:",omitempty"`
	InterfaceCount                 int           `json:",omitempty"`
	InterfaceNameMap               struct{}      `json:",omitempty"`
	LicenseCount                   int           `json:",omitempty"`
	Interfaces                     []struct{}    `json:",omitempty"`
	WebEtherIP                     string        `json:",omitempty"`
	IPSupport                      string        `json:",omitempty"`
	DefaultGateway                 interface{}   `json:",omitempty"`
	ClusterPrivateEtherIP          string        `json:",omitempty"`
	NullValueColumnNames           []interface{} `json:",omitempty"`
	EnableSwitchInService          bool          `json:",omitempty"`
	UserInterfaceIPForCP           string        `json:",omitempty"`
	NetworkInterfaceIpsUnique      bool          `json:",omitempty"`
}

type ControllerUtilization struct {
	Cbutils []struct {
		QuarterStats []struct {
			CPUPerc          string `json:",omitempty"`
			QuarterStartTime string `json:",omitempty"`
			MemoryPerc       string `json:",omitempty"`
			DiskPerc         string `json:",omitempty"`
			DiskTotal        string `json:",omitempty"`
			DiskFree         string `json:",omitempty"`
		} `json:",omitempty"`
		ControlID string `json:",omitempty"`
	} `json:",omitempty"`
	Timestamp string `json:",omitempty"`
}

type SystemSummary struct {
	DpVersion                             string      `json:",omitempty"`
	NumOfControlBlades                    int         `json:",omitempty"`
	NumOfDisconnectedAPs                  int         `json:",omitempty"`
	CpVersion                             string      `json:",omitempty"`
	FailoverEnable                        bool        `json:",omitempty"`
	BladeName                             string      `json:",omitempty"`
	DataCollectingEnable                  bool        `json:",omitempty"`
	BladeState                            string      `json:",omitempty"`
	NumOfInServiceControlBlades           int         `json:",omitempty"`
	ClusterOperationBlockAfterUpgradeFail bool        `json:",omitempty"`
	NumOfOutOfServiceDataBlades           int         `json:",omitempty"`
	ControllerLicenseSummary              string      `json:",omitempty"`
	NumOfConnectedAPsByRadioTypeMap       interface{} `json:",omitempty"`
	NumOfOutOfServiceControlBlades        int         `json:",omitempty"`
	ZoneIPMode                            string      `json:",omitempty"`
	NumOfDataBlades                       int         `json:",omitempty"`
	ClusterName                           string      `json:",omitempty"`
	Model                                 string      `json:",omitempty"`
	BladeInServiceMap                     struct{}    `json:",omitempty"`
	ClusterState                          string      `json:",omitempty"`
	NodeAffinityEnable                    bool        `json:",omitempty"`
	ClusterType                           string      `json:",omitempty"`
	SerialNumber                          string      `json:",omitempty"`
	VdpExternalVirtualCount               string      `json:",omitempty"`
	IPAddress                             string      `json:",omitempty"`
	IPSupport                             string      `json:",omitempty"`
	Version                               string      `json:",omitempty"`
	VdpLicenseSummary                     string      `json:",omitempty"`
	ClusterRedundancyType                 string      `json:",omitempty"`
	Uptime                                int         `json:",omitempty"`
	NumOfInServiceDataBlades              int         `json:",omitempty"`
	MaxSupportedAps                       string      `json:",omitempty"`
	ClusterRedundancyRole                 string      `json:",omitempty"`
	VdpExternalPhysicalCount              string      `json:",omitempty"`
	ApVersion                             string      `json:",omitempty"`
	DomainName                            interface{} `json:",omitempty"`
	GatewayLicenseSummary                 string      `json:",omitempty"`
	NumOfConnectedAPs                     int         `json:",omitempty"`
	RxgwlicenseSummary                    string      `json:",omitempty"`
}
