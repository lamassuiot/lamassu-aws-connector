package consul

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/lamassuiot/lamassu-aws-connector/pkg/server/discovery"
	uuid "github.com/satori/go.uuid"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

type ServiceDiscovery struct {
	client       consulsd.Client
	logger       log.Logger
	registration *api.AgentServiceRegistration
}

func NewServiceDiscovery(consulProtocol string, consulHost string, consulPort string, CA string, logger log.Logger) (discovery.Service, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulProtocol + "://" + consulHost + ":" + consulPort
	tlsConf := &api.TLSConfig{CAFile: CA}
	consulConfig.TLSConfig = *tlsConf
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not start Consul API Client")
		return nil, err
	}
	client := consulsd.NewClient(consulClient)
	return &ServiceDiscovery{client: client, logger: logger}, nil
}

func (sd *ServiceDiscovery) Register(advProtocol string, advPort string, tags []string, name string, persistenceDir string) (string, error) {

	cmd := exec.Command("hostname", "-i")
	stdout, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	strStdout := strings.TrimSuffix(string(stdout), "\n")

	check := api.AgentServiceCheck{
		HTTP:          advProtocol + "://" + strStdout + ":" + advPort + "/v1/health",
		Interval:      "10s",
		Timeout:       "1s",
		TLSSkipVerify: true,
		Notes:         "Basic health checks",
	}

	meta := map[string]string{"connector-type": tags[0], "name": name}

	var svcId string

	if _, err := os.Stat(persistenceDir + "/identifier"); errors.Is(err, os.ErrNotExist) {
		svcId = uuid.NewV4().String()
		os.WriteFile(persistenceDir+"/identifier", []byte(svcId), 0640)
	} else {
		data, err := os.ReadFile(persistenceDir + "/identifier")
		if err != nil {
			return "", nil
		}
		svcId = string(data)
		isValidUUID := func(uuid string) bool {
			r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
			return r.MatchString(uuid)
		}
		if !isValidUUID(svcId) {
			return "", errors.New("invalid uuidv4 string")
		}
	}
	level.Info(sd.logger).Log("msg", "connector ID is: "+svcId)
	port, _ := strconv.Atoi(advPort)
	asr := api.AgentServiceRegistration{
		ID:      svcId,
		Name:    "cloud-connector",
		Address: strStdout,
		Port:    port,
		Tags:    tags,
		Meta:    meta,
		Check:   &check,
	}
	sd.registration = &asr
	return svcId, sd.client.Register(sd.registration)
}

func (sd *ServiceDiscovery) Deregister() error {
	return sd.client.Deregister(sd.registration)
}
