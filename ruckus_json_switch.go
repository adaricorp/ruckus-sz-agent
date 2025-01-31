package main

type SwitchPowerSupplyStatus struct {
	SerialNumber        string `json:",omitempty"`
	PowerSupplySlotList []struct {
		SlotNumber int    `json:",omitempty"`
		Type       string `json:",omitempty"`
		Status     string `json:",omitempty"`
		Unit       int    `json:",omitempty"`
	} `json:",omitempty"`
}

type SwitchFanStatus struct {
	SerialNumber string `json:",omitempty"`
	FanSlotList  []struct {
		SlotNumber int         `json:",omitempty"`
		Type       interface{} `json:",omitempty"`
		Status     string      `json:",omitempty"`
	} `json:",omitempty"`
}

type SwitchTemperatureStatus struct {
	StackID             string `json:",omitempty"`
	SerialNumber        string `json:",omitempty"`
	TemperatureSlotList []struct {
		SlotNumber       int     `json:",omitempty"`
		TemperatureValue float64 `json:",omitempty"`
	} `json:",omitempty"`
}

type VLANDetail struct {
	VlanID   string `json:",omitempty"`
	VlanName string `json:",omitempty"`
}
