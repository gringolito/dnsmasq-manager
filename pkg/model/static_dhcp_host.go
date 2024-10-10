package model

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
)

type StaticDhcpHost struct {
	MacAddress net.HardwareAddr
	IPAddress  net.IP
	HostName   string
}

const errInvalidDHCPHostConfig = "invalid DHCP host config: %s"

var ErrDHCPHostMissingMACAddress = errors.New("invalid DHCP host: missing MAC address")
var ErrDHCPHostMissingIPAddress = errors.New("invalid DHCP host: missing IP address")
var ErrDHCPHostMissingHostName = errors.New("invalid DHCP host: missing hostname")

func (h *StaticDhcpHost) FromConfig(config string) error {
	tokens := strings.Split(config, ",")
	if len(tokens) != 3 {
		return fmt.Errorf(errInvalidDHCPHostConfig, config)
	}

	var mac string
	_, err := fmt.Sscanf(tokens[0], "dhcp-host=%s", &mac)
	if err != nil {
		return errors.Join(fmt.Errorf(errInvalidDHCPHostConfig, config), err)
	}

	h.MacAddress, err = net.ParseMAC(mac)
	h.IPAddress = net.ParseIP(tokens[1])
	if h.IPAddress == nil {
		err = errors.Join(err, &net.AddrError{Err: "invalid IP address", Addr: tokens[1]})
	}

	h.HostName = tokens[2]

	return err
}

func (h *StaticDhcpHost) check() error {
	var err error = nil
	if h.MacAddress.String() == "" {
		err = errors.Join(err, ErrDHCPHostMissingMACAddress)
	}
	if h.IPAddress.String() == "<nil>" {
		err = errors.Join(err, ErrDHCPHostMissingIPAddress)
	}
	if h.HostName == "" {
		err = errors.Join(err, ErrDHCPHostMissingHostName)
	}
	return err
}

func (h *StaticDhcpHost) ToConfig() (string, error) {
	err := h.check()
	if err != nil {
		return "", err
	}

	config := fmt.Sprintf("dhcp-host=%s,%s,%s", h.MacAddress.String(), h.IPAddress.String(), h.HostName)
	return config, nil
}

func (h *StaticDhcpHost) Equal(other StaticDhcpHost) bool {
	return bytes.Equal(h.MacAddress, other.MacAddress) && bytes.Equal(h.IPAddress, other.IPAddress) && h.HostName == other.HostName
}
