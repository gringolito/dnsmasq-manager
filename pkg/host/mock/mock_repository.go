package hostmock

import (
	"net"

	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
}

func (m *RepositoryMock) Delete(host *model.StaticDhcpHost) (*model.StaticDhcpHost, error) {
	args := m.Called(host)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) DeleteByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) DeleteByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	args := m.Called(ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) Find(host *model.StaticDhcpHost) (*model.StaticDhcpHost, error) {
	args := m.Called(host)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) FindAll() (*[]model.StaticDhcpHost, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) FindByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) FindByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	args := m.Called(ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StaticDhcpHost), args.Error(1)
}

func (m *RepositoryMock) Save(host *model.StaticDhcpHost) error {
	args := m.Called(host)
	return args.Error(0)
}
