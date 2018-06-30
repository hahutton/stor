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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hahutton/stor/providers"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var dryRun bool
var recurse bool

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:   "cp [//alias/]source_file... [//alias/]target_file",
	Short: "Copy blobs between providers with cp like semantics",
	Long: `Copy blobs between cloud storage and/or the local filesystem.

cp utilizes an alias for different providers. The alias is set in the config
file. It contains necessary info to connect to cloud storage accounts such as account
names and keys.

Local file system [//file/]source_file... can be a file name(s) or directories. If directory and 
not -R then it will be skipped. The typical shell expands * before passing it to any cmd. stor handles this
on the local file provider by handling multiple sources with the last positional arg being the target.

Object store //alias/source_file can be a file name or a prefix.
The match semantics are specific to the cloud providers.

The prefix semantics match the substring of characters at the beginning of the key.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		argCount := len(args)
		targetPosition := argCount - 1

		sourceAlias, sourcePathName := providers.Parse(args[0])
		targetAlias, targetPathName := providers.Parse(args[targetPosition])
		jww.INFO.Printf("sourceAlias: %s, sourcePathName: %s", sourceAlias, sourcePathName)
		jww.INFO.Printf("targetAlias: %s, targetPathName: %s", targetAlias, targetPathName)

		sourceProvider := providers.Create(sourceAlias)
		targetProvider := providers.Create(targetAlias)

		var sourceInfos []*providers.BlobInfo
		for _, arg := range args[:targetPosition] {
			jww.INFO.Println("arg:", arg)
			statInfo := sourceProvider.Stat(arg)
			if !statInfo.IsDir {
				sourceInfos = append(sourceInfos, statInfo)
			} else {
				if recurse {
					filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
							return err
						}
						if info.Mode().IsRegular() {
							blobInfo := &providers.BlobInfo{}
							blobInfo.Name = info.Name()
							blobInfo.PathName = path
							blobInfo.IsDir = info.IsDir()
							blobInfo.Length = info.Size()
							blobInfo.LastModified = info.ModTime()
							sourceInfos = append(sourceInfos, blobInfo)
						}
						return nil
					})
				}
			}
		}

		if dryRun {
			for _, sourceInfo := range sourceInfos {
				fmt.Printf("%s\n", sourceInfo.PathName)
			}
			os.Exit(0)
		}

		for _, sourceInfo := range sourceInfos {

			blockCount, blockSize := providers.CalculateBlocks(sourceInfo)

			var targetName string

			if isDir(targetPathName) {
				targetName = fmt.Sprintf("%s%s", targetPathName, sourceInfo.PathName)
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
	cpCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "show set of blobs to be copied but don't copy")
	cpCmd.Flags().BoolVarP(&recurse, "Recurse", "R", false, "Recurse directories mainly for local file provider")
}
