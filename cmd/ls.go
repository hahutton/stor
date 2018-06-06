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
	//"fmt"
	"os"

	//"github.com/hahutton/stor/blob"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	//"github.com/spf13/viper"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls //alias[/pathname]",
	Short: "ls (list) blobs in an azure storage container",
	Long:  `list the contents of an azure blob storage container or a subset thereof.`,
	Run: func(cmd *cobra.Command, args []string) {

		if !isTargetAliased(args[0]) {
			jww.ERROR.Println("stor ls expects a storage account alias in the form of //<alias_name>.")
			os.Exit(1)
		}

		jww.TRACE.Println("ls called")
	},
}

func init() {
	RootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
