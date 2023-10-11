package utils

type Config struct {
	ApplicationName        string
	ClientName             string
	CadenceService         string
	HostPort               string
	Domain                 string
	TaskListName           string
	CadenceFrontendService string
	CadenceClientService   string
}

func GetConfig() *Config {
	return &Config{
		ApplicationName:        "cadence-test",
		ClientName:             "simpleworker",
		CadenceService:         "cadence-service",
		HostPort:               "127.0.0.1:7933",
		Domain:                 "cadence-test",
		TaskListName:           "SimpleWorker",
		CadenceFrontendService: "cadence-frontend",
		CadenceClientService:   "cadence-client",
	}
}
