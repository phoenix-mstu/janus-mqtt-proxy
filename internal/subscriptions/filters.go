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
}

func MakeFilter(topicFilter regexp.Regexp, topicAssembler template.Template, valMap map[string]template.Template) Filter {
	return Filter{
		brokerTopicFilter:    topicFilter,
		clientTopicAssembler: topicAssembler,
		valueMap:             valMap,
	}
}

func (filter *Filter) Apply(topic string, value []byte) (string, []byte, bool) {
	match := filter.brokerTopicFilter.FindStringSubmatch(topic)
	if len(match) < 2 {
		return "", nil, false
	}

	args := map[string]string{}
	for i := 0; i < len(match) - 1; i++ {
		args[fmt.Sprintf("f%v", i + 1)] = match[i + 1]
	}
	clientTopic := executeTemplate(filter.clientTopicAssembler, args)

	clientValue := string(value)
	if len(filter.valueMap) > 0 {
		if newValue, ok := filter.valueMap[clientValue]; ok {
			clientValue = executeTemplate(newValue, args)
		} else {
			return "", nil, false
		}
	}

	return clientTopic, []byte(clientValue), true
}

func executeTemplate(t template.Template, args map[string]string) string {
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, args); err != nil {
		panic(err)
	}
	return strings.Replace(buf.String(), "\n"," ",-1)
}