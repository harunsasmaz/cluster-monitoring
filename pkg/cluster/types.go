package cluster

import (
	"fmt"
	"strings"
)

type LabelParams struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (p *LabelParams) AsFilter() string {
	return fmt.Sprintf("%s=%s", p.Label, p.Value)
}

func toLabelSelector(params ...LabelParams) string {
	var filters []string
	for _, param := range params {
		filters = append(filters, param.AsFilter())
	}

	return strings.Join(filters, ",")
}

type ServiceInfo struct {
	Name             string `json:"name"`
	ApplicationGroup string `json:"applicationGroup"`
	RunningPodsCount int    `json:"runningPodsCount"`
}
