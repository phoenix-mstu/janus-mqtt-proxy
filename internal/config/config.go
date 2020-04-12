package config

import (
	"errors"
	"github.com/phoenix-mstu/go-modifying-mqtt-proxy/internal/subscriptions"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"regexp"
	"text/template"
)

type CompiledFiltersConfig struct {
	Login string
	Password string
	BrokerFilters []subscriptions.Filter
	ClientFilters []subscriptions.Filter
}

type CompiledConfig struct {
	BrokerHost string
	BrokerLogin string
	BrokerPassword string
	Clients []CompiledFiltersConfig
}

func ReadConfigFile(filename string) *CompiledConfig {
	yamlMainConfig := YamlMainConfig{}

	if err := yaml.Unmarshal(readFile(filename), &yamlMainConfig); err != nil {
		log.Fatalf("Can't decode yaml %v: %v", filename, err.Error())
	}

	res := CompiledConfig{
		BrokerHost:     yamlMainConfig.BrokerHost,
		BrokerLogin:    yamlMainConfig.BrokerLogin,
		BrokerPassword: yamlMainConfig.BrokerPassword,
	}

	for _, cConfig := range yamlMainConfig.Clients {
		yamlFiltersConfig := YamlFiltersConfig{}

		if err := yaml.Unmarshal(readFile(cConfig.FiltersConfig), &yamlFiltersConfig); err != nil {
			log.Fatalf("Can't decode yaml %v: %v", filename, err.Error())
		}

		if len(yamlFiltersConfig.ClientFilters) == 0 && len(yamlFiltersConfig.BrokerFilters) == 0 {
			log.Fatalf("Error in %v: either broker_to_client or client_to_broker must not be empty", cConfig.FiltersConfig)
		}

		bf, err := ReadFiltersFromConfig(yamlFiltersConfig.BrokerFilters)
		if err != nil {
			log.Fatalf("Error in %v: %v", cConfig.FiltersConfig, err)
		}

		cf, err := ReadFiltersFromConfig(yamlFiltersConfig.ClientFilters)
		if err != nil {
			log.Fatalf("Error in %v: %v", cConfig.FiltersConfig, err)
		}

		res.Clients = append(res.Clients, CompiledFiltersConfig{
			Login:         cConfig.Login,
			Password:      cConfig.Password,
			BrokerFilters: bf,
			ClientFilters: cf,
		})
	}

	if res.BrokerHost == "" {
		log.Fatal("Missing broker_host")
	}

	if len(res.Clients) == 0 {
		log.Fatal("Empty clients list")
	}

	return &res
}

func readFile(filename string) []byte {
	file, err := ioutil.ReadFile(filename)
	if  err != nil {
		log.Fatalf("Can't read config file %v: %v", filename, err.Error())
	}
	return file
}

func ReadFiltersFromConfig(filtersConfig []YamlFilterConfig) ([]subscriptions.Filter, error) {
	var filters []subscriptions.Filter
	for _, filterConfig := range filtersConfig {
		if filterConfig.Topic == "" || filterConfig.Template == "" {
			return nil, errors.New("both topic and template must not be empty")
		}

		valuesMap := map[string]template.Template{}
		for key, value := range filterConfig.ValMap {
			valuesMap[key] = *makeTemplate(value)
		}

		var valFilter *regexp.Regexp
		var valAssembler *template.Template
		if filterConfig.ValRegex != "" &&  filterConfig.ValTemplate != "" {
			valFilter = regexp.MustCompile(filterConfig.ValRegex)
			valAssembler = makeTemplate(filterConfig.ValTemplate)
		}

		filters = append(filters, subscriptions.MakeFilter(
			*regexp.MustCompile(filterConfig.Topic),
			*makeTemplate(filterConfig.Template),
			valuesMap,
			valFilter,
			valAssembler))
	}
	return filters, nil
}

func makeTemplate(s string) *template.Template {
	t, _ := template.New("").Parse(s)
	return t
}
