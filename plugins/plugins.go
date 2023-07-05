package plugins

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/file"
)

const (
	goBinEnvVar      = "GOBIN"
	visionSeparator  = "-"
	visionFirstWord  = "vision"
	visionSecondWord = "plugin"
)

func GetPlugins(executor execute.Executor) ([]string, error) {
	var plugins []string
	pluginPath, err := goBinPath(executor)
	if err != nil {
		return plugins, err
	}
	pluginFiles, err := file.ReadDir(pluginPath)
	if err != nil {
		return plugins, fmt.Errorf("cannot read plugin directory %s: %s", pluginPath, err.Error())
	}
	for _, pluginFile := range pluginFiles {
		if !pluginFile.IsDir() && fileIsVisionPlugin(pluginFile.Name()) {
			plugins = append(plugins, pluginFile.Name())
		}
	}
	return plugins, nil
}

func goBinPath(executor execute.Executor) (string, error) {
	goBinPath := file.GetEnv(goBinEnvVar)
	if goBinPath == "" {
		goPath, err := executor.Output(exec.Command("go", "env", "GOPATH"), ".", "getting GOPATH")
		if err != nil {
			return "", err
		}
		goBinPath = string(goPath)[:len(goPath)-1] + "/bin"
	}
	return goBinPath, nil
}

func fileIsVisionPlugin(filename string) bool {
	c := strings.Split(filename, visionSeparator)
	if len(c) != 4 || c[0] != visionFirstWord || c[1] != visionSecondWord {
		return false
	}
	return true
}
