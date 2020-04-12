package subscriptions

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

type Filter struct {
	brokerTopicFilter regexp.Regexp
	clientTopicAssembler template.Template
	valueMap map[string]template.Template
	valueFilter *regexp.Regexp
	valueAssembler *template.Template
}

func MakeFilter(
	topicFilter regexp.Regexp,
	topicAssembler template.Template,
	valMap map[string]template.Template,
	valFilter *regexp.Regexp,
	valAssembler *template.Template) Filter {
	return Filter{
		brokerTopicFilter:    topicFilter,
		clientTopicAssembler: topicAssembler,
		valueMap:             valMap,
		valueFilter: 		  valFilter,
		valueAssembler:       valAssembler,
	}
}

func (filter *Filter) Apply(topic string, value []byte) (string, []byte, bool) {
	match := filter.brokerTopicFilter.FindStringSubmatch(topic)
	if len(match) < 1 {
		return "", nil, false
	}

	args := prepareTemplateArgs(match)
	clientTopic := executeTemplate(filter.clientTopicAssembler, args)

	clientValue := string(value)
	if len(filter.valueMap) > 0 {
		if newValue, ok := filter.valueMap[clientValue]; ok {
			clientValue = executeTemplate(newValue, args)
		} else {
			return "", nil, false
		}
	} else if filter.valueFilter != nil && filter.valueAssembler != nil {
		match := filter.valueFilter.FindStringSubmatch(clientValue)
		if len(match) < 2 {
			return "", nil, false
		}
		args := prepareTemplateArgs(match)
		clientValue = executeTemplate(*filter.valueAssembler, args)
	}



	return clientTopic, []byte(clientValue), true
}

func prepareTemplateArgs(match []string) map[string]string {
	args := map[string]string{}
	for i := 0; i < len(match) - 1; i++ {
		args[fmt.Sprintf("f%v", i + 1)] = match[i + 1]
	}
	return args
}

func executeTemplate(t template.Template, args map[string]string) string {
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, args); err != nil {
		panic(err)
	}
	return strings.Replace(buf.String(), "\n"," ",-1)
}