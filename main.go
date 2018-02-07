package main

import (
	"gopkg.in/yaml.v2"
	"bufio"
	"os"
	"path"
	"github.com/sirupsen/logrus"
	"fmt"
	"gopkg.in/validator.v2"
	"strings"
	"io/ioutil"
)

type ApplicationYaml struct {
	User string `yaml:"user"`
	InfraFilePath string `yaml:"infra_file"`
}

type SSHConfig struct {
	Env string `validate:"nonzero"`
	User string `validate:"nonzero"`
	Config []struct{
		HostAlias string `yaml:"hostAlias" validate:"nonzero"`
		Host  string `validate:"nonzero"`
	}
}

var sshConfigDataSample = `
- env: s
  user: ubuntu
  config:
    - hostAlias: token-service-01
      host: 10.14.1.214
    - hostAlias: voucher-service-01
      host: 10.14.1.182

- env: p
  user: ${user}
  config:
    - hostAlias: token-service-01
      host: 10.20.1.214
`

func main() {
	app := GetConfigEnvUser("Application.yaml")
	logrus.Info("configured user: ", app.User)

	infraFileContent := GetFileContent(app.InfraFilePath)

	yoConfigContent := GetYoConfigContent(string(infraFileContent))

	aliases := GetAliases(yoConfigContent, app.User)
	logrus.Info(aliases)

	userHomeDir := GetUserHomeDir()
	yoConfigPath := path.Join(userHomeDir, ".yo_config")
	yoConfigSourceStr := fmt.Sprintf("source %s", yoConfigPath)

	CreateYoConfig(yoConfigPath, aliases)

	bashProfilePath := GetBashProfilePath(userHomeDir)
	if ! IsSSHConfigSourcedInBashProfile(bashProfilePath, yoConfigSourceStr) {
		AppendYoSourceToBashProfile(bashProfilePath, yoConfigSourceStr)
	}
	logrus.Infof("run source %s or open new terminal to activate aliases", bashProfilePath)
}

func GetFileContent(filePath string) []byte  {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Fatal(err)
	}
	return b
}

func GetConfigEnvUser(filePath string) ApplicationYaml {
	b := GetFileContent(filePath)
	var applicationYaml ApplicationYaml
	err := yaml.Unmarshal(b, &applicationYaml)

	if err != nil {
		logrus.Fatalf("error yaml.Unmarshal: %v", err)
	}

	return applicationYaml
}

func GetYoConfigContent(yamlContent string) []SSHConfig  {
	sshConfig := []SSHConfig{}
	e := yaml.Unmarshal([]byte(yamlContent), &sshConfig)
	if e != nil {
		logrus.Fatalf("error yaml.Unmarshal: %v", e)
	}
	return sshConfig
}

func GetAliases(sshConfig []SSHConfig, configUser string) string  {
	var aliases []string

	for i:=0; i<len(sshConfig); i++ {
		sshConfigItem := sshConfig[i]

		if errs := validator.Validate(sshConfigItem); errs != nil {
			logrus.Error("not valid: ", sshConfigItem)
			logrus.Fatalf(errs.Error())
		}

		configItems := sshConfigItem.Config
		for j := 0; j < len(configItems); j++ {
			configItem := configItems[j]
			user := sshConfigItem.User
			if user == "${user}" {
				user = configUser
			}
			aliases = append(aliases, fmt.Sprintf("alias %s=\"ssh %s@%s\"", configItem.HostAlias, user, configItem.Host) )
		}
	}
	return strings.Join(aliases, "\n")
}

func CreateYoConfig(yoConfigPath string, yoConfigContent string)  {
	err := os.Remove(yoConfigPath)
	if err != nil {
		logrus.Info(err)
	}

	logrus.Info("Creating yo_config in user home ", yoConfigPath)

	f, err := os.Create(yoConfigPath)
	if err != nil {
		logrus.Fatal(err)
	}
	if _, err := f.WriteString(yoConfigContent); err != nil {
		logrus.Fatal(err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		logrus.Fatal(err)
	}
	if err := f.Close(); err != nil {
		logrus.Fatal(err)
	}
}

func AppendYoSourceToBashProfile(bashfilePath string, yoSourceStr string)  {
	logrus.Info("Appending yo_config to bash_profile")
	f, err := os.OpenFile(bashfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatal(err)
	}
	if _, err := f.Write([]byte(yoSourceStr)); err != nil {
		logrus.Fatal(err)
	}
	if err := f.Close(); err != nil {
		logrus.Fatal(err)
	}
}

func GetBashProfilePath(userHomeDir string) string {
	bashProfile := path.Join(userHomeDir,".bash_profile")
	logrus.Info("bash profile: ",bashProfile)
	return bashProfile
}

func GetUserHomeDir() string {
	home := os.Getenv("HOME")

	logrus.Info( "user home: ",home )
	return home
}

func IsSSHConfigSourcedInBashProfile(bashfilePath string, yoConfigPath string) bool  {
	file, err := os.Open(bashfilePath)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		bashProfileLine := scanner.Text()
		if bashProfileLine == yoConfigPath {
			logrus.Info("yo_config entry exists in bash_profile")
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Fatal(err)
	}
	return false

}
