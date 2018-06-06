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
	"net/url"
	"os"
	"time"

	"github.com/hahutton/stor/blob"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:   "cp source destination",
	Short: "copy source destination",
	Long: `cp copies files and/or directories to and from 
Azure Storage`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		lastCount := len(args) - 1
		if isTargetAliased(args[lastCount]) {
			alias, blobName, _ := parse(args[lastCount])

			url, _ := url.Parse(viper.GetString(fmt.Sprintf("aliases.%s.url", alias)))

			container := &blob.Container{
				Alias:       alias,
				AccountName: viper.GetString(fmt.Sprintf("aliases.%s.accountName", alias)),
				Name:        viper.GetString(fmt.Sprintf("aliases.%s.name", alias)),
				Endpoint:    url,
				Key:         viper.GetString(fmt.Sprintf("aliases.%s.key", alias)),
			}

			jww.TRACE.Println("Container Create")
			jww.TRACE.Println("Container.Alias", container.Alias)
			jww.TRACE.Println("Container.AccountName", container.AccountName)
			jww.TRACE.Println("Container.Name", container.Name)
			jww.TRACE.Println("Container.Endpoint", container.Endpoint)
			jww.TRACE.Println("Container.Key", container.Key)

			//TODO Make blockSize dynamic???
			viper.SetDefault("blockSize", 4*1024*1024)
			blockSize := viper.GetInt("blockSize")

			for _, fileName := range args[0:lastCount] {

				fileInfo, _ := os.Stat(fileName)
				if !fileInfo.IsDir() {
					transfer := container.TransferIn(fileInfo.Name(), fileInfo.Size(), blobName, blockSize)
					resp, err := transfer.Do()
					if err != nil {
						jww.ERROR.Println(err)
						os.Exit(1)
					}
					jww.INFO.Println(resp)
				}
			}
			duration := time.Since(start)
			jww.INFO.Printf("Elapsed: %v\n", duration)
		} else {
			jww.ERROR.Println("only copy to azure implemented")
		}
	},
}

func init() {
	RootCmd.AddCommand(cpCmd)

	//TODO handle more files at once
	//cpCmd.Flags().BoolP("recurse", "r", false, "Recursively add files to source")
}

