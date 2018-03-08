package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// ServerCommand
func InitFileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "init config",
		RunE: func(cmd *cobra.Command, args []string) error {

			d, _ := yaml.Marshal(cfg())

			dir := cfg().HomePath
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					panic(fmt.Sprintf("Could not create directory %v. %v", dir, err))
				}
			}

			if err := ioutil.WriteFile(path.Join(cfg().HomePath, "config.yaml"), d, 0755); err != nil {
				panic(err.Error())
			}

			return nil
		},
	}
}
