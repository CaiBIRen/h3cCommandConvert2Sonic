package model

type Sonic_response struct {
	Code  int16
	Value string
}

type Mac_interface struct {
	Ifname string
	Mac    string
}

type Mac_interface_list struct {
	Mac_interfaces []Mac_interface
}
