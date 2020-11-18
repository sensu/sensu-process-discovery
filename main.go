package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/shirou/gopsutil/process"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	SubPrefix string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-process-discovery",
			Short:    "Discover system processes and output a list of agent subscriptions.",
			Keyspace: "sensu.io/plugins/sensu-process-discovery/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "subscription-prefix",
			Env:       "SUBSCRIPTION_PREFIX",
			Argument:  "subscription-prefix",
			Shorthand: "p",
			Default:   "",
			Usage:     "The agent subscription name prefix",
			Value:     &plugin.SubPrefix,
		},
	}
)

var subMap = map[string]string{
	"sensu-backend":  "sensu-backend",
	"node_exporter":  "node-exporter",
	"postgres":       "postgres",
	"apache":         "apache",
	"monit$":         "monit",
	"httpd":          "apache",
	"couchdb":        "couchdb",
	"etcd":           "etcd",
	"haproxy":        "haproxy",
	"mongod":         "mongodb",
	"openvpn":        "openvpn",
	"fluentd":        "fluentd",
	"jenkins":        "jenkins",
	"redis":          "redis",
	"varnish":        "varnish",
	"cassandra":      "cassandra",
	"hbase":          "hbase",
	"kafka":          "kafka",
	"mysql":          "mysql",
	"resque":         "resque",
	"sidekiq":        "sidekiq",
	"syslog":         "syslog",
	"vsphere":        "vsphere",
	"ceph":           "ceph",
	"kubernetes":     "kubernetes",
	"nginx":          "nginx",
	"qmgr":           "postfix",
	"pickup":         "postfix",
	"rethinkdb":      "rethinkdb",
	"docker":         "docker",
	"gitlab":         "gitlab",
	"iis":            "iis",
	"lxc":            "lxc",
	"ntp":            "ntp",
	"chronyd":        "chrony",
	"riak":           "riak",
	"tomcat":         "tomcat",
	"consul":         "consul",
	"elasticsearch":  "elasticsearch",
	"gluster":        "gluster",
	"solr":           "solr",
	"tripwire":       "tripwire",
	"memcached":      "memcached",
	"openldap":       "openldap",
	"rabbitmq":       "rabbitmq",
	"spark":          "spark",
	"zookeeper":      "zookeeper",
	"couchbase":      "couchbase",
	"unicorn":        "unicorn",
	"salt-master":    "salt-master",
	"salt-minion":    "salt-minion",
	"smbd":           "samba",
	"grafana-server": "grafana",
	"influxd":        "influxdb",
}

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	return sensu.CheckStateOK, nil
}

func processSubs() ([]string, error) {
	subs := []string{}
	subsSet := make(map[string]bool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	plist, err := process.ProcessesWithContext(ctx)

	if err != nil {
		return subs, err
	}

	for _, p := range plist {
		n, err := p.NameWithContext(ctx)

		if err != nil {
			continue
		}

		for r, s := range subMap {
			// Not sure if we want or need regex. Regex
			// makes more sense when dealing with the full
			// process list line (could even match and
			// extract specific arguments, i.e. ports).
			m, err := regexp.Match(r, []byte(n))

			if err != nil {
				continue
			}

			if m {
				if _, e := subsSet[s]; !e {
					subs = append(subs, plugin.SubPrefix+s)
					subsSet[s] = true
				}
			}

		}
	}

	return subs, nil
}

func executeCheck(event *corev2.Event) (int, error) {
	subs, err := processSubs()

	fmt.Println(strings.Join(subs, "\n"))

	if err != nil {
		return sensu.CheckStateWarning, err
	}

	return sensu.CheckStateOK, nil
}
