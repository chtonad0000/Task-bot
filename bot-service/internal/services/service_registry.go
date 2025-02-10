package services

type ServiceRegistry struct {
	services map[string]ServiceClient
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{services: make(map[string]ServiceClient)}
}

func (r *ServiceRegistry) RegisterService(s ServiceClient) {
	r.services[s.GetName()] = s
}

func (r *ServiceRegistry) GetService(name string) (ServiceClient, bool) {
	s, ok := r.services[name]
	return s, ok
}
