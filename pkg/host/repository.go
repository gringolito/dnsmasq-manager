package host

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"golang.org/x/exp/slog"
)

type Repository interface {
	Delete(host *model.StaticDhcpHost) (*model.StaticDhcpHost, error)
	DeleteByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error)
	DeleteByIP(ipAddress net.IP) (*model.StaticDhcpHost, error)
	Find(host *model.StaticDhcpHost) (*model.StaticDhcpHost, error)
	FindAll() (*[]model.StaticDhcpHost, error)
	FindByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error)
	FindByIP(ipAddress net.IP) (*model.StaticDhcpHost, error)
	Save(host *model.StaticDhcpHost) error
}

type repository struct {
	staticHostsFilePath string
	mutex               sync.RWMutex
}

func NewRepository(staticHostsFilePath string) Repository {
	return &repository{
		staticHostsFilePath: staticHostsFilePath,
	}
}

func (r *repository) FindAll() (*[]model.StaticDhcpHost, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.load()
}

func (r *repository) Find(host *model.StaticDhcpHost) (*model.StaticDhcpHost, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.find(sameHost(host))
}

func (r *repository) FindByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.find(sameMacAddress(macAddress))
}

func (r *repository) FindByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.find(sameIPAddress(ipAddress))
}

func (r *repository) Save(host *model.StaticDhcpHost) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	hosts, err := r.load()
	if err != nil {
		return err
	}

	*hosts = append(*hosts, *host)
	return r.save(hosts)
}

func (r *repository) Delete(host *model.StaticDhcpHost) (*model.StaticDhcpHost, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.delete(sameHost(host))
}

func (r *repository) DeleteByMac(macAddress net.HardwareAddr) (*model.StaticDhcpHost, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.delete(sameMacAddress(macAddress))
}

func (r *repository) DeleteByIP(ipAddress net.IP) (*model.StaticDhcpHost, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.delete(sameIPAddress(ipAddress))
}

func (r *repository) load() (*[]model.StaticDhcpHost, error) {
	file, err := os.Open(r.staticHostsFilePath)
	if err != nil {
		slog.Error("Error reading static hosts file",
			slog.String("file", r.staticHostsFilePath),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer file.Close()

	return r.parse(file)
}

func (r *repository) parse(file *os.File) (*[]model.StaticDhcpHost, error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	hosts := []model.StaticDhcpHost{}
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "dhcp-host=") {
			slog.Debug("Skipping line", slog.String("line", scanner.Text()))
			continue
		}
		slog.Debug("Parsing line", slog.String("line", scanner.Text()))

		host := model.StaticDhcpHost{}
		err := host.FromConfig(scanner.Text())
		if err != nil {
			slog.Error("Failed to parse static DHCP host entry",
				slog.String("entry", scanner.Text()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}

		hosts = append(hosts, host)
	}

	return &hosts, nil
}

func (r *repository) save(hosts *[]model.StaticDhcpHost) error {
	config := make([]string, 0, len(*hosts))
	for _, host := range *hosts {
		hostConfig, err := host.ToConfig()
		if err != nil {
			return err
		}
		config = append(config, hostConfig)
	}

	err := os.WriteFile(r.staticHostsFilePath, []byte(strings.Join(config, "\n")), os.FileMode(0644))
	if err != nil {
		slog.Error("Error writing into the static hosts file",
			slog.String("file", r.staticHostsFilePath),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (r *repository) delete(filter Filter) (*model.StaticDhcpHost, error) {
	hosts, err := r.load()
	if err != nil {
		return nil, err
	}

	h := *hosts
	for i, host := range h {
		if !filter(host) {
			continue
		}

		*hosts = append(h[:i], h[i+1:]...)
		err := r.save(hosts)
		return &host, err
	}

	return nil, nil
}

func (r *repository) find(filter Filter) (*model.StaticDhcpHost, error) {
	hosts, err := r.load()
	if err != nil {
		return nil, err
	}

	for _, host := range *hosts {
		if filter(host) {
			return &host, nil
		}
	}

	return nil, nil
}

type Filter func(model.StaticDhcpHost) bool

func sameHost(host *model.StaticDhcpHost) Filter {
	return func(other model.StaticDhcpHost) bool {
		return host.Equal(other)
	}
}
func sameMacAddress(macAddress net.HardwareAddr) Filter {
	return func(other model.StaticDhcpHost) bool {
		return bytes.Equal(macAddress, other.MacAddress)
	}
}

func sameIPAddress(ipAddress net.IP) Filter {
	return func(other model.StaticDhcpHost) bool {
		return bytes.Equal(ipAddress, other.IPAddress)
	}
}
