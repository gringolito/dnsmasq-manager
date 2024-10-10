package model

import (
	"fmt"
	"net"
	"testing"

	"github.com/gringolito/dnsmasq-manager/tests"
	"github.com/stretchr/testify/assert"
)

const (
	ValidHostConfig            = `dhcp-host=02:04:06:aa:bb:cc,1.1.1.1,Foo`
	InvalidMacAddressConfig    = `dhcp-host=ab:cd:ef:gh:ij:kl,1.1.1.1,Jung`
	InvalidIPAddressConfig     = `dhcp-host=02:04:06:aa:bb:cc,11.1.1,Jung`
	InvalidBothAddressesConfig = `dhcp-host=ab:cd:ef:gh:ij:kl,11.1.1,Jung`
	InvalidConfig              = `not-dhcp-config`
	InvalidConfig2             = `02:04:06:aa:bb:cc,1.1.1.1,Jung`
	MissingMacAddressConfig    = `dhcp-host=1.1.1.1,Foo`
	MissingIPAddressConfig     = `dhcp-host=02:04:06:aa:bb:cc,Foo`
	MissingHostNameConfig      = `dhcp-host=02:04:06:aa:bb:cc,1.1.1.1`
	InvalidIPAddress           = `11.1.1`
	InvalidMacAddress          = `ab:cd:ef:gh:ij:kl`
)

var ValidHost = StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"}

func TestStaticDhcpHostFromConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config string
		assert func(t *testing.T, host *StaticDhcpHost, err error)
	}{
		{
			name:   "Success",
			config: ValidHostConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.NoError(t, err, "StaticDhcpHost.FromConfig() returned an unexpected error")
				assert.Equal(t, host, &ValidHost, "StaticDhcpHost.FromConfig() has generated an unexpected host")
			},
		},
		{
			name:   "InvalidIPAddress",
			config: InvalidIPAddressConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.EqualError(t, err, fmt.Sprintf("address %s: invalid IP address", InvalidIPAddress), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "InvalidMacAddress",
			config: InvalidMacAddressConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.EqualError(t, err, fmt.Sprintf("address %s: invalid MAC address", InvalidMacAddress), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "InvalidBothAddresses",
			config: InvalidBothAddressesConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.ErrorContains(t, err, fmt.Sprintf("address %s: invalid MAC address", InvalidMacAddress), "StaticDhcpHost.FromConfig() returned an unexpected error")
				assert.ErrorContains(t, err, fmt.Sprintf("address %s: invalid IP address", InvalidIPAddress), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "NotADhcpHost",
			config: InvalidConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.EqualError(t, err, fmt.Sprintf(errInvalidDHCPHostConfig, InvalidConfig), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "NotADhcpHost2",
			config: InvalidConfig2,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.ErrorContains(t, err, fmt.Sprintf(errInvalidDHCPHostConfig, InvalidConfig2), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "MissingMacAddress",
			config: MissingMacAddressConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.EqualError(t, err, fmt.Sprintf(errInvalidDHCPHostConfig, MissingMacAddressConfig), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "MissingIPAddress",
			config: MissingIPAddressConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.EqualError(t, err, fmt.Sprintf(errInvalidDHCPHostConfig, MissingIPAddressConfig), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
		{
			name:   "MissingHostName",
			config: MissingHostNameConfig,
			assert: func(t *testing.T, host *StaticDhcpHost, err error) {
				assert.Error(t, err, "StaticDhcpHost.FromConfig() did NOT returned error")
				assert.EqualError(t, err, fmt.Sprintf(errInvalidDHCPHostConfig, MissingHostNameConfig), "StaticDhcpHost.FromConfig() returned an unexpected error")
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			host := StaticDhcpHost{}
			err := host.FromConfig(test.config)
			test.assert(t, &host, err)
		})
	}
}

func TestStaticDhcpHostToConfig(t *testing.T) {
	testCases := []struct {
		name   string
		host   StaticDhcpHost
		assert func(t *testing.T, config string, err error)
	}{
		{
			name: "Success",
			host: ValidHost,
			assert: func(t *testing.T, config string, err error) {
				assert.NoError(t, err, "StaticDhcpHost.ToConfig() returned an unexpected error")
				assert.Equal(t, ValidHostConfig, config, "StaticDhcpHost.ToConfig() returned an unexpected config string")
			},
		},
		{
			name: "MissingMacAddress",
			host: StaticDhcpHost{IPAddress: net.ParseIP("1.1.1.1"), HostName: "FooBar"},
			assert: func(t *testing.T, config string, err error) {
				assert.Error(t, err, "StaticDhcpHost.ToConfig() did NOT returned an error")
				assert.ErrorIs(t, err, ErrDHCPHostMissingMACAddress, "StaticDhcpHost.ToConfig returned an unexpected error")
			},
		},
		{
			name: "MissingIPAddress",
			host: StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:ab:cd:ef"), HostName: "FooBar"},
			assert: func(t *testing.T, config string, err error) {
				assert.Error(t, err, "StaticDhcpHost.ToConfig() did NOT returned an error")
				assert.ErrorIs(t, err, ErrDHCPHostMissingIPAddress, "StaticDhcpHost.ToConfig returned an unexpected error")
			},
		},
		{
			name: "MissingHostName",
			host: StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:ab:cd:ef"), IPAddress: net.ParseIP("1.1.1.1")},
			assert: func(t *testing.T, config string, err error) {
				assert.Error(t, err, "StaticDhcpHost.ToConfig() did NOT returned an error")
				assert.ErrorIs(t, err, ErrDHCPHostMissingHostName, "StaticDhcpHost.ToConfig returned an unexpected error")
			},
		},
		{
			name: "EmptyHost",
			assert: func(t *testing.T, config string, err error) {
				assert.Error(t, err, "StaticDhcpHost.ToConfig() did NOT returned an error")
				assert.ErrorIs(t, err, ErrDHCPHostMissingMACAddress, "StaticDhcpHost.ToConfig returned an unexpected error")
				assert.ErrorIs(t, err, ErrDHCPHostMissingIPAddress, "StaticDhcpHost.ToConfig returned an unexpected error")
				assert.ErrorIs(t, err, ErrDHCPHostMissingHostName, "StaticDhcpHost.ToConfig returned an unexpected error")
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			config, err := test.host.ToConfig()
			test.assert(t, config, err)
		})
	}
}

func TestStaticDhcpHostEqual(t *testing.T) {
	testCases := []struct {
		name   string
		a      StaticDhcpHost
		b      StaticDhcpHost
		result bool
	}{
		{
			name:   "EmptyHosts",
			a:      StaticDhcpHost{},
			b:      StaticDhcpHost{},
			result: true,
		},
		{
			name:   "SameHosts",
			a:      ValidHost,
			b:      ValidHost,
			result: true,
		},
		{
			name:   "DifferentIpAddresses",
			a:      StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
			b:      StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.2"), HostName: "Foo"},
			result: false,
		},
		{
			name:   "DifferentMacAddresses",
			a:      StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
			b:      StaticDhcpHost{MacAddress: tests.ParseMAC("12:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
			result: false,
		},
		{
			name:   "DifferentHostnames",
			a:      StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
			b:      StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Bar"},
			result: false,
		},
		{
			name:   "AllDifferent",
			a:      StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
			b:      StaticDhcpHost{MacAddress: tests.ParseMAC("12:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.2"), HostName: "Bar"},
			result: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.a.Equal(test.b))
		})
	}
}
