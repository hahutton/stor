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
	"time"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var cpCmd = &cobra.Command{
	Use:   "cp source destination",
	Short: "copy source destination",
	Long:  `cp copies files and/or directories to and from Azure Storage`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		src := args[0]
		target := args[1]

		if isAlias(src) && !isAlias(target) {
			jww.ERROR.Println("cp implemented for push only currently")
			os.Exit(1)
		}

		if isGlob(src) && !isDir(target) {
			jww.ERROR.Printf("Globbed src must have dir target. %s -> %s\n", src, target)
			os.Exit(1)
		}

		fileNames, err := filepath.Glob(src)
		if err != nil {
			jww.ERROR.Printf("Bad pattern for files: %s", src)
			jww.ERROR.Println(err)
			os.Exit(1)
		}
		jww.TRACE.Println("FileNames:", fileNames)

		container, pathName, blockSize := parse(target)

		for _, fileName := range fileNames {

			fileInfo, _ := os.Stat(fileName)

			var blobName string
			//TODO do you want to flatten it??
			//fileName = fileInfo.Name()

			if isDir(pathName) {
				blobName = fmt.Sprintf("%s%s", pathName, fileName)
			} else {
				blobName = pathName
			}
			transfer := container.TransferIn(fileName, fileInfo.Size(), blobName, blockSize)
			resp, err := transfer.Do()
			if err != nil {
				jww.ERROR.Println(err)
				os.Exit(1)
			}
			jww.INFO.Println(resp)
		}
		duration := time.Since(start)
		jww.INFO.Printf("Elapsed: %v\n", duration)
	},
}

func init() {
	RootCmd.AddCommand(cpCmd)
}
