package chart

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

// Changelog return any information available for a helm chart
func (c *Chart) Changelog(name string) (string, error) {
	URL := fmt.Sprintf("%s/index.yaml", c.URL)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	index, err := loadIndex(body)

	if err != nil {
		return "", err
	}

	e, err := index.Get(c.Name, c.Version)

	if err != nil {
		return "", err
	}

	t := &template.Template{}
	t = template.Must(template.New("changelog").Parse(CHANGELOGTEMPLATE))
	buffer := new(bytes.Buffer)

	type params struct {
		Name        string
		Description string
		Home        string
		KubeVersion string
		Created     string
		URL         []string
		SOURCES     []string
	}

	err = t.Execute(buffer, params{
		Name:        e.Name,
		Description: e.Description,
		Home:        e.Home,
		KubeVersion: e.KubeVersion,
		Created:     e.Created.String(),
		URL:         e.URLs,
		SOURCES:     e.Sources})

	if err != nil {
		return "", err
	}

	changelog := buffer.String()

	return changelog, nil

}
