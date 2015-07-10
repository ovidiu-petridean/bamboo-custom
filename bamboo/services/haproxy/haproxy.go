package haproxy

import (
	"github.com/QubitProducts/bamboo/Godeps/_workspace/src/github.com/samuel/go-zookeeper/zk"
	conf "github.com/QubitProducts/bamboo/configuration"
	"github.com/QubitProducts/bamboo/services/marathon"
	"github.com/QubitProducts/bamboo/services/service"
	"fmt"
)

type templateData struct {
	Apps     marathon.AppList
	Services map[string]service.Service
}

func GetTemplateData(config *conf.Configuration, conn *zk.Conn) interface{} {

	apps, _ := marathon.FetchApps(config.Marathon)
	services, _ := service.All(conn, config.Bamboo.Zookeeper)


	fmt.Println("aaaaaaaaaaa")
	fmt.Println(apps)
	fmt.Println("bbbbbbbbbbb")

	return templateData{apps, services}
}
