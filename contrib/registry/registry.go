package registry

type Registry struct {
	services []*ServiceEntry
}

type ServiceEntry struct {
	name     string
	endpoint string
}

func (e *ServiceEntry) Name() string {
	return e.name
}

func (e *ServiceEntry) Endpoint() string {
	return e.endpoint
}
