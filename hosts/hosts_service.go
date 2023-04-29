package hosts

type hostService struct {
	repository hostRepository
}

func NewService(staticHostsFilePath string) hostService {
	return hostService{repository: NewRepository(staticHostsFilePath)}
}

func (s *hostService) getStaticHosts() ([]staticDhcpHost, error) {
	return s.repository.Load()
}

func (s *hostService) addStaticHost(host staticDhcpHost) error {
	return s.repository.Insert(host)
}
