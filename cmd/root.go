// Copyright © 2018 Hays Hutton <hays.hutton@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var cfgFile string
var VeryVerbose bool
var Verbose bool

var RootCmd = &cobra.Command{
	Use:   "stor",
	Short: "An Azure Blob Storage utility",
	Long: `stor is a cli tool to interact with azure storage.
While stor aims to be a sharp tool with a more unix philosophy,
azcopy should be used whenever possible due to its robustness and
feature set. stor aims to have no dependencies which is a difference.`,
}

// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		jww.ERROR.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.stor.yml then $HOME/.stor.yml)")
	RootCmd.PersistentFlags().BoolVarP(&VeryVerbose, "trace", "t", false, "very verbose")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if Verbose && !VeryVerbose {
		jww.SetStdoutThreshold(jww.LevelInfo)
	}

	if VeryVerbose {
		jww.SetStdoutThreshold(jww.LevelTrace)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			jww.ERROR.Println(home)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName(".stor")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		jww.ERROR.Println("Bad config file:", viper.ConfigFileUsed())
		jww.ERROR.Println("Bad config file:", err)
	}
}
