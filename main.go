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
	UserNames     []UsersNames `yaml:"user_names"`
	InfraFilePath string       `yaml:"infra_file"`
}

type UsersNames struct {
	UserToken string `yaml: "usertoken" validate:"nonzero"`
	UserName string `yaml: "username" validate:"nonzero"`
}

type SSHConfig struct {
	Env      string `validate:"nonzero"`
	UserName string `yaml:"user_name" validate:"nonzero"`
	Config []struct{
		HostAlias string `yaml:"host_alias" validate:"nonzero"`
		Host  string `validate:"nonzero"`
	}
}

func main() {
	appYamlPath := "Application.yaml"

	if len(os.Args) == 2 {
		appYamlPath = os.Args[1]
		logrus.Infof("Using config yaml from path: %s", appYamlPath)
	}

	app := GetConfigEnvUser(appYamlPath)
	logrus.Info("configured users: ", app.UserNames)

	infraFileContent := GetFileContent(app.InfraFilePath)

	yoConfigContent := GetYoConfigContent(string(infraFileContent))

	aliases := createAliases(yoConfigContent, app.UserNames)
	//logrus.Info(aliases)
	//logrus.Info("yoConfigContent")

	userHomeDir := GetUserHomeDir()
	yoConfigPath := path.Join(userHomeDir, ".yo_config")
	yoConfigSourceStr := fmt.Sprintf("source %s", yoConfigPath)

	CreateYoConfig(yoConfigPath, aliases)

	bashProfilePath := GetBashProfilePath(userHomeDir)
	if ! IsYoConfigSourcedInFile(bashProfilePath, yoConfigSourceStr) {
		AppendYoSourceToFile(bashProfilePath, yoConfigSourceStr)
	}

	zshrcPath := GetZshrcPath(userHomeDir)
	if ! IsYoConfigSourcedInFile(zshrcPath, yoConfigSourceStr) {
		AppendYoSourceToFile(zshrcPath, yoConfigSourceStr)
	}

	logrus.Infof("Open new terminal to activate aliases", bashProfilePath)
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

func createAliases(sshConfig []SSHConfig, usersNames []UsersNames) string  {
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
			userName := findUserName(usersNames, sshConfigItem.UserName)
			aliases = append(aliases, fmt.Sprintf("alias %s=\"ssh %s@%s\"", configItem.HostAlias, userName, configItem.Host) )
		}
	}
	return strings.Join(aliases, "\n")
}

func findUserName(userNames []UsersNames, userToken string) string  {
	for _,v := range userNames {
		if fmt.Sprintf("${%s}", v.UserToken) == userToken {
			return v.UserName
		}
	}
	panic(fmt.Sprintf("no user token defined in application.yaml for user: %s", userToken))
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

func AppendYoSourceToFile(bashfilePath string, yoSourceStr string)  {
	logrus.Infof("Appending yo_config to %s", bashfilePath)
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
	filePath := path.Join(userHomeDir,".bash_profile")
	logrus.Info("bash profile: ", filePath)
	return filePath
}

func GetZshrcPath(userHomeDir string) string {
	filePath := path.Join(userHomeDir,".zshrc")
	logrus.Info(".zshrc profile: ", filePath)
	return filePath
}

func GetUserHomeDir() string {
	home := os.Getenv("HOME")

	logrus.Info( "user home: ",home )
	return home
}

func IsYoConfigSourcedInFile(bashfilePath string, yoConfigPath string) bool  {
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
