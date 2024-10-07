package hostmock

import (
	"net"

	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) Insert(host *model.StaticDhcpHost) error {
	args := m.Called(host)
	return args.Error(0)
}

func (m *ServiceMock) Update(host *model.StaticDhcpHost) error {
	args := m.Called(host)
	return args.Error(0)
}

func (m *ServiceMock) FetchAll() (*[]model.StaticDhcpHost, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.StaticDhcpHost), args.Error(1)
}

func (m *ServiceMock) FetchByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	args := m.Called(ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *ServiceMock) FetchByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *ServiceMock) RemoveByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	args := m.Called(ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *ServiceMock) RemoveByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}
