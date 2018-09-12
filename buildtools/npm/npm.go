package npm

import (
	"encoding/json"
	"os"

	"github.com/apex/log"
	"github.com/fossas/fossa-cli/exec"
	"github.com/fossas/fossa-cli/files"
	"github.com/mitchellh/mapstructure"
)

type Options struct {
	AllowNPMErr bool `mapstructure:"allow-npm-err"`
}

type NPM interface {
	List(dir string) (Output, error)
	Clean(dir string) error
	Install(dir string) error
}

type SystemNPM struct {
	Cmd      string
	AllowErr bool
}

type Output struct {
	Version      string
	From         string
	Resolved     string
	Dependencies map[string]Output
}

func (n SystemNPM) List(dir string) (Output, error) {
	stdout, _, err := exec.Run(exec.Cmd{
		Name: n.Cmd,
		Argv: []string{"ls", "--json"},
		Dir:  dir,
	})
	log.Debugf("err: %#v", err)
	log.Debugf("AllowErr: %#v", n.AllowErr)
	if err != nil && !n.AllowErr {
		return Output{}, err
	}
	var output Output
	err = json.Unmarshal([]byte(stdout), &output)
	if err != nil {
		return Output{}, err
	}
	return output, nil
}

func (n SystemNPM) Clean(dir string) error {
	return files.Rm(dir, "node_modules")
}

func (n SystemNPM) Install(dir string) error {
	_, _, err := exec.Run(exec.Cmd{
		Name: n.Cmd,
		Argv: []string{"install", "--production"},
		Dir:  dir,
	})
	if err != nil && !n.AllowErr {
		return err
	}
	return nil
}

func New(options map[string]interface{}) NPM {
	npmCmd, _, npmErr := exec.Which("-v", os.Getenv("FOSSA_NPM_CMD"), "npm")

	if npmErr != nil {
		log.Warnf("Could not find NPM: %s", npmErr.Error())
	}

	var decodedOptions Options
	mapstructure.Decode(options, &decodedOptions)

	return SystemNPM{
		Cmd:      npmCmd,
		AllowErr: decodedOptions.AllowNPMErr || true,
	}
}
