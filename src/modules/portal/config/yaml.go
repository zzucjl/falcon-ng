package config

import (
	"fmt"
	"sync"

	"github.com/toolkits/pkg/file"
)

// PortalYml -> etc/portal.yml
type PortalYml struct {
	Salt   string            `yaml:"salt"`
	Logger loggerSection     `yaml:"logger"`
	HTTP   httpSection       `yaml:"http"`
	LDAP   ldapSection       `yaml:"ldap"`
	Proxy  proxySection      `yaml:"proxy"`
	Judges map[string]string `yaml:"judges"`
}

type loggerSection struct {
	Dir       string `yaml:"dir"`
	Level     string `yaml:"level"`
	KeepHours uint   `yaml:"keepHours"`
}

type httpSection struct {
	Listen string `yaml:"listen"`
	Secret string `yaml:"secret"`
}

type ldapSection struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	BaseDn     string `yaml:"baseDn"`
	BindUser   string `yaml:"bindUser"`
	BindPass   string `yaml:"bindPass"`
	AuthFilter string `yaml:"authFilter"`
	TLS        bool   `yaml:"tls"`
	StartTLS   bool   `yaml:"startTLS"`
}

type proxySection struct {
	Transfer string `yaml:"transfer"`
	Index    string `yaml:"index"`
}

var (
	yaml *PortalYml
	lock = new(sync.RWMutex)
)

// Get configuration file
func Get() *PortalYml {
	lock.RLock()
	defer lock.RUnlock()
	return yaml
}

// Parse configuration file
func Parse(ymlfile string) error {
	var c PortalYml
	err := file.ReadYaml(ymlfile, &c)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", ymlfile, err)
	}

	lock.Lock()
	yaml = &c
	lock.Unlock()

	return nil
}
