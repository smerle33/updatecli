package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition check if a specific chart version exist
func (c *Chart) Condition(source string) (bool, error) {
	if c.spec.Version != "" {
		logrus.Infof("Version %v, already defined from configuration file", c.spec.Version)
	} else {
		c.spec.Version = source
	}
	URL := fmt.Sprintf("%s/index.yaml", c.spec.URL)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	index, err := loadIndex(body)
	if err != nil {
		return false, err
	}

	message := ""
	if c.spec.Version != "" {
		message = fmt.Sprintf(" for version '%s'", c.spec.Version)
	}

	if index.Has(c.spec.Name, c.spec.Version) {
		logrus.Infof("%s Helm Chart '%s' is available on %s%s", result.SUCCESS, c.spec.Name, c.spec.URL, message)
		return true, nil
	}

	logrus.Infof("%s Helm Chart '%s' isn't available on %s%s", result.FAILURE, c.spec.Name, c.spec.URL, message)
	return false, nil
}

// ConditionFromSCM returns an error because it's not supported
func (c *Chart) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Helm chart condition, aborting")
}
