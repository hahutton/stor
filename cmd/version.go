// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"fmt"

	"github.com/spf13/cobra"
)

var PrimaryVersion string = "0.2"
var Version string
var MasterRev string
var BranchRev string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version information",
	Long: `Version information which is revision counts on master.branch.  Plus revision SHA1s for
master and branch. If branch revision exists then this binary is a development build. The master
revisions match to https://github.com/hahutton/stor revisions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("stor Block Storage Command Tool v%s.%s\n", PrimaryVersion, Version)
		fmt.Printf("master revision %s\n", MasterRev)
		fmt.Printf("branch revision %s\n", BranchRev)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
