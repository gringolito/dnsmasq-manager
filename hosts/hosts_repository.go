package hosts

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type hostRepository struct {
	staticHostsFilePath string
}

func NewRepository(staticHostsFilePath string) hostRepository {
	return hostRepository{staticHostsFilePath: staticHostsFilePath}
}

func (r *hostRepository) Load() ([]staticDhcpHost, error) {
	file, err := os.Open(r.staticHostsFilePath)
	if err != nil {
		log.Printf("Error reading static hosts file (%s): %s", r.staticHostsFilePath, err.Error())
		return nil, err
	}
	defer file.Close()

	return r.parse(file)
}

func (r *hostRepository) parse(file *os.File) ([]staticDhcpHost, error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	hosts := []staticDhcpHost{}
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "dhcp-host=") {
			log.Println("Skipping line:", scanner.Text())
			continue
		}
		log.Println("Parsing line:", scanner.Text())

		host := staticDhcpHost{}
		err := host.FromConfig(scanner.Text())
		if err != nil {
			log.Printf("Failed to parse static DHCP host entry (%s): %s", scanner.Text(), err.Error())
			return nil, err
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}