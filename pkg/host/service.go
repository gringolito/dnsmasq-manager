package host

import (
	"fmt"

	"github.com/gringolito/pi-hole-manager/pkg/model"
)

type Service interface {
	Insert(host *model.StaticDhcpHost) error
	Update(host *model.StaticDhcpHost) error
	FetchAll() (*[]model.StaticDhcpHost, error)
	FetchByIP(ipAddress string) (*model.StaticDhcpHost, error)
	FetchByMac(macAddress string) (*model.StaticDhcpHost, error)
	RemoveByIP(ipAddress string) (*model.StaticDhcpHost, error)
	RemoveByMac(macAddress string) (*model.StaticDhcpHost, error)
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
		return fmt.Errorf("Duplicated MAC address")
	}

	sameIPHost, err := s.repository.FindByIP(host.IPAddress)
	if err != nil {
		return err
	}
	if sameIPHost != nil {
		return fmt.Errorf("Duplicated IP address")
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

func (s *service) FetchByMac(macAddress string) (*model.StaticDhcpHost, error) {
	return s.repository.FindByMac(macAddress)
}

func (s *service) FetchByIP(ipAddress string) (*model.StaticDhcpHost, error) {
	return s.repository.FindByIP(ipAddress)
}

func (s *service) RemoveByMac(macAddress string) (*model.StaticDhcpHost, error) {
	return s.repository.DeleteByMac(macAddress)
}

func (s *service) RemoveByIP(ipAddress string) (*model.StaticDhcpHost, error) {
	return s.repository.DeleteByIP(ipAddress)
}
