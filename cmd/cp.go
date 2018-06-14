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
	"strings"
	"time"

	"github.com/hahutton/stor/providers"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:   "cp //alias/source //alias/target",
	Short: "stor cp [//alias/]source_file [//alias/]target_file",
	Long:  `provider cp`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		sourceAlias, sourcePathName := providers.Parse(args[0])
		targetAlias, targetPathName := providers.Parse(args[1])
		jww.INFO.Printf("sourceAlias: %s, sourcePathName: %s", sourceAlias, sourcePathName)
		jww.INFO.Printf("targetAlias: %s, targetPathName: %s", targetAlias, targetPathName)

		sourceProvider := providers.Create(sourceAlias)
		targetProvider := providers.Create(targetAlias)

		sourceInfos := sourceProvider.Glob(sourcePathName)

		for _, sourceInfo := range sourceInfos {

			blockCount, blockSize := providers.CalculateBlocks(sourceInfo)

			var targetName string

			if isDir(targetPathName) {
				targetName = fmt.Sprintf("%s%s", targetPathName, sourceInfo.Name)
			} else {
				targetName = targetPathName
			}

			tokenBucket := providers.InitTokenBucket()

			//starting with blockCount here since that shouldn't backpressure at all
			//might want to use this to govern memory usage at some point??
			//at that point introduce config var to "dial" this
			transferChan := make(chan *providers.Block, blockCount) //TODO this could kill on memory. fast big read
			sourceProvider.Open(sourceInfo.PathName, transferChan, tokenBucket, blockCount, blockSize)
			jww.INFO.Println(targetName)
			targetProvider.Create(targetName, transferChan, blockCount, tokenBucket)
		}
		duration := time.Since(start)
		jww.INFO.Printf("Elapsed: %v\n", duration)
	},
}

func isDir(path string) bool {
	return strings.HasSuffix(path, "/")
}

func init() {
	RootCmd.AddCommand(cpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
