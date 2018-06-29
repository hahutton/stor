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
	"fmt"
	"time"

	"github.com/hahutton/stor/providers"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var Long bool
var NoHeader bool

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [//alias/]source_name [flags]",
	Short: "List blobs",
	Long:  `List blobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		sourceAlias, sourcePathName := providers.Parse(args[0])

		sourceProvider := providers.Create(sourceAlias)
		sourceInfos := sourceProvider.Glob(sourcePathName)

		if !NoHeader {
			fmt.Printf("%s\n", sourcePathName)
		}

		var layout string = "Jan 02 15:04"
		for _, si := range sourceInfos {
			if Long {
				fmt.Printf("%s  %s %s %10d %s\n", si.BlobType, si.LastModified.Format(layout), si.Etag, si.Length, si.Name)
			} else {
				fmt.Println(si.Name)
			}
		}

		duration := time.Since(start)
		jww.INFO.Printf("Elapsed: %v\n", duration)
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
	lsCmd.Flags().BoolVarP(&Long, "long", "l", false, "included extended attributes")
	lsCmd.Flags().BoolVarP(&NoHeader, "noheader", "n", false, "remove header from output")
}
