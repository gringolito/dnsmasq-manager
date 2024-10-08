package host

import (
	"fmt"
	"net"

	"github.com/gringolito/dnsmasq-manager/pkg/model"
)

type Service interface {
	Insert(host *model.StaticDhcpHost) error
	Update(host *model.StaticDhcpHost) error
	FetchAll() (*[]model.StaticDhcpHost, error)
	FetchByIP(ipAddress net.IP) (*model.StaticDhcpHost, error)
	FetchByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error)
	RemoveByIP(ipAddress net.IP) (*model.StaticDhcpHost, error)
	RemoveByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) Insert(host *model.StaticDhcpHost) error {
	sameMacHost, err := s.repository.FindByMac(host.MacAddress)
	if err != nil {
		return err
	}
	if sameMacHost != nil {
		return &DuplicatedEntryError{Field: "MAC", Value: host.MacAddress.String()}
	}

	sameIPHost, err := s.repository.FindByIP(host.IPAddress)
	if err != nil {
		return err
	}
	if sameIPHost != nil {
		return &DuplicatedEntryError{Field: "IP", Value: host.IPAddress.String()}
	}

	return s.repository.Save(host)
}

func (s *service) Update(host *model.StaticDhcpHost) error {
	_, err := s.repository.DeleteByMac(host.MacAddress)
	if err != nil {
		return err
	}

	_, err = s.repository.DeleteByIP(host.IPAddress)
	if err != nil {
		return err
	}

	return s.repository.Save(host)
}

func (s *service) FetchAll() (*[]model.StaticDhcpHost, error) {
	return s.repository.FindAll()
}

func (s *service) FetchByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	return s.repository.FindByMac(macAddress)
}

func (s *service) FetchByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	return s.repository.FindByIP(ipAddress)
}

func (s *service) RemoveByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	return s.repository.DeleteByMac(macAddress)
}

func (s *service) RemoveByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	return s.repository.DeleteByIP(ipAddress)
}

type DuplicatedEntryError struct {
	Field string
	Value string
}

const duplicatedEntryErrorMessage = "Duplicated %s address: %s"

func (e DuplicatedEntryError) Error() string {
	return fmt.Sprintf(duplicatedEntryErrorMessage, e.Field, e.Value)
}
