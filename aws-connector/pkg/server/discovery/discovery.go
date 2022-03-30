package discovery

type Service interface {
	Register(advProtocol string, advPort string, tags []string, name string, persistenceDir string) (string, error)
	Deregister() error
}
