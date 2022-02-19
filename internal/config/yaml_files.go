package config

type YamlFilterConfig struct {
	Topic string
	Template string
	ValMap map[string]string `yaml:"val_map"`
	ValRegex string `yaml:"val_regex"`
	ValTemplate string `yaml:"val_template"`
}

type YamlFiltersConfig struct {
	BrokerFilters []YamlFilterConfig `yaml:"client_to_broker"`
	ClientFilters []YamlFilterConfig `yaml:"broker_to_client"`
}

type YamlMainConfig struct {
	BrokerHost string `yaml:"broker_host"`
	BrokerClientID string `yaml:"broker_client_id"`
	BrokerLogin string `yaml:"broker_login"`
	BrokerPassword string `yaml:"broker_password"`
	Clients []struct{
		ClientID string `yaml:"client_id"`
		Login string
		Password string
		FiltersConfig string `yaml:"filters_config"`
	}
}
