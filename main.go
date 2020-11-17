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
	Example string
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
		&sensu.PluginConfigOption{
			Path:      "example",
			Env:       "CHECK_EXAMPLE",
			Argument:  "example",
			Shorthand: "e",
			Default:   "",
			Usage:     "An example string configuration option",
			Value:     &plugin.Example,
		},
	}
)

var subMap = map[string]string{
	".*sensu-backend.*": "sensu-backend",
	"postgres.*":        "postgres",
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
			m, err := regexp.Match(r, []byte(n))

			if err != nil {
				continue
			}

			if m {
				if _, e := subsSet[s]; !e {
					subs = append(subs, s)
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
