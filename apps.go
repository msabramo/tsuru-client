package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/timeredbull/tsuru/cmd"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type AppGrant struct{}

func (c *AppGrant) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "app-grant",
		Usage:   "app-grant <appname> <teamname>",
		Desc:    "grants access to an app to a team.",
		MinArgs: 2,
	}
}

func (c *AppGrant) Run(context *cmd.Context, client cmd.Doer) error {
	appName, teamName := context.Args[0], context.Args[1]
	url := cmd.GetUrl(fmt.Sprintf("/apps/%s/%s", appName, teamName))
	request, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		return err
	}
	io.WriteString(context.Stdout, fmt.Sprintf(`Team "%s" was added to the "%s" app`+"\n", teamName, appName))
	return nil
}

type AppRevoke struct{}

func (c *AppRevoke) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "app-revoke",
		Usage:   "app-revoke <appname> <teamname>",
		Desc:    "revokes access to an app from a team.",
		MinArgs: 2,
	}
}

func (c *AppRevoke) Run(context *cmd.Context, client cmd.Doer) error {
	appName, teamName := context.Args[0], context.Args[1]
	url := cmd.GetUrl(fmt.Sprintf("/apps/%s/%s", appName, teamName))
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		return err
	}
	io.WriteString(context.Stdout, fmt.Sprintf(`Team "%s" was removed from the "%s" app`+"\n", teamName, appName))
	return nil
}

type AppModel struct {
	Name  string
	State string
	Units []Units
}

type Units struct {
	Ip string
}

type AppList struct{}

func (c *AppList) Run(context *cmd.Context, client cmd.Doer) error {
	request, err := http.NewRequest("GET", cmd.GetUrl("/apps"), nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return c.Show([]byte(result), context)
}

func (c *AppList) Show(result []byte, context *cmd.Context) error {
	var apps []AppModel
	err := json.Unmarshal(result, &apps)
	if err != nil {
		return err
	}
	table := cmd.NewTable()
	table.Headers = cmd.Row([]string{"Application", "State", "Ip"})
	for _, app := range apps {
		ip := ""
		if len(app.Units) > 0 {
			ip = app.Units[0].Ip
		}
		table.AddRow(cmd.Row([]string{app.Name, app.State, ip}))
	}
	context.Stdout.Write(table.Bytes())
	return nil
}

func (c *AppList) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "app-list",
		Usage: "app-list",
		Desc:  "list all your apps.",
	}
}

type AppCreate struct{}

func (c *AppCreate) Run(context *cmd.Context, client cmd.Doer) error {
	appName := context.Args[0]
	framework := context.Args[1]

	b := bytes.NewBufferString(fmt.Sprintf(`{"name":"%s", "framework":"%s"}`, appName, framework))
	request, err := http.NewRequest("POST", cmd.GetUrl("/apps"), b)
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	out := make(map[string]string)
	err = json.Unmarshal(result, &out)
	if err != nil {
		return err
	}
	io.WriteString(context.Stdout, fmt.Sprintf(`App "%s" successfully created!`+"\n", appName))
	io.WriteString(context.Stdout, fmt.Sprintf(`Your repository for "%s" project is "%s"`, appName, out["repository_url"])+"\n")
	return nil
}

func (c *AppCreate) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "app-create",
		Usage:   "app-create <appname> <framework>",
		Desc:    "create a new app.",
		MinArgs: 2,
	}
}

type AppRemove struct{}

func (c *AppRemove) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "app-remove",
		Usage:   "app-remove <appname>",
		Desc:    "removes an app.",
		MinArgs: 1,
	}
}

func (c *AppRemove) Run(context *cmd.Context, client cmd.Doer) error {
	appName := context.Args[0]
	url := cmd.GetUrl(fmt.Sprintf("/apps/%s", appName))
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		return err
	}
	io.WriteString(context.Stdout, fmt.Sprintf(`App "%s" successfully removed!`+"\n", appName))
	return nil
}

type AppLog struct{}

func (c *AppLog) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "log",
		Usage:   "log <appname>",
		Desc:    "show logs for an app.",
		MinArgs: 1,
	}
}

type Log struct {
	Date    time.Time
	Message string
}

func (c *AppLog) Run(context *cmd.Context, client cmd.Doer) error {
	appName := context.Args[0]
	url := cmd.GetUrl(fmt.Sprintf("/apps/%s/log", appName))
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	logs := []Log{}
	err = json.Unmarshal(result, &logs)
	if err != nil {
		return err
	}
	for _, log := range logs {
		context.Stdout.Write([]byte(log.Date.String() + " - " + log.Message + "\n"))
	}
	return err
}
