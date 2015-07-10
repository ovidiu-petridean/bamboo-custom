package configuration

import (
	"fmt"
	"strings"
)

/*
	Mesos Marathon configuration
*/
type Marathon struct {
	// comma separated marathon http endpoints including port number
	Endpoint string
}

func (m Marathon) Endpoints() []string {
	fmt.Println(strings.Split(m.Endpoint, ","))
	return strings.Split(m.Endpoint, ",")
}
