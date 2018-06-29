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

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)


const skeleton_config_file string = `
# Config for stor

# Each aliases sub key is an alias for a storage account / container combo which is in the url
# plus its storage account key
# This is just a convenience method to shorten what is required to type
# The key itself is the alias to basically the url/key pair for the storage account
#   e.g.  //blah/filename   -> https://hhblah.blob.core.windows.net/blah/filename
#         plus it references the right "key" too so none of that has to be passed
#    or   //stor/downloads/macosx/stor -> https://hahutton.blob.core.windows.net/stor/downloads/macosx/stor
#   This means one could type: stor cp ./filename //blah/filename
#    to copy a file from the local filesystem to the Azure Storage account 
#   The net effect of this is to use "//blah/" instead of "https://hhblah.blob.core.windows.net/blah/"
#   which really is a lot less typing (plus the associated key will be used too!
#
# Once you fill out the below. It is most convenient to move it to your $HOME directory. That way you can
# use stor from anywhere. It will also check for .stor.yml in the current working directory. Plus, there is
# a command switch to pass it on the cmd line too.
# Example of a prior config (with invalid keys now :)
# ---------------------------------------------------
#	aliases:
#     blah: 
#       provider: azure
#       name: blah
#       accountName: hhblah
#       key: kY5kQiy8ECepzfJUjC2H47Lt8RR/B091nagd8uRf+Hfa8oiuut8jV8uN2krOEM9WuaO9nBBvCv7rK9fDW6b8Bg== 
#     stor:
#       provider: azure
#       name: stor
#       accountName: hahutton
#       key: TA+RADLa2PfySyyPjEaUp+P8LRXxKeRxEBqQLnRunbwZggKNPCoQD5zTxWJ4OkHMoKD2o7CfRPrNTc/05x5GTQ==
#     file:
#       provider: file


aliases:
  <your_alias_here>: 
    provider: azure
    name: <your_container_name_here>
    accountName: <your_storage_account_name_here>
    key: <your storage_account_key here>
  <another_alias_here>:
    provider: azure
    name: <your_container_name_here> 
    accountName: <your_storage_account_name_here>
    key: <your storage_account_key here>
  file:                                        #This must be here if you plan to use the local filesystem
    provider: file

# The blockSize defines the size of the chunks a file or blob is split into. The parallelism * blockSize is
# a direct correlate with memory consumption. Azure Storage used to have a 4MB size. Currently it can range from
# 0 to 100MB per block. Tweak for you workload.

#blockSize: 5242880   #5MB
blockSize: 10485760  #10MB
#blockSize: 20971520  #20MB
`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a skeleton config file",
	Long: `init creates a skeleton config file.
	
stor maps multiple values under aliases in a config file. stor can 
point to many providers such as local file system, Azure Blob Storage,
eventually AWS S3 and maybe even Google Cloud Storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat(".stor.yml")
		if err == nil {
			jww.ERROR.Println(".stor.yml already exists in current directory. ")
			jww.ERROR.Println("Remove first in order to create new one. ")
			os.Exit(1)
		}

		configFile, err := os.Create(".stor.yml")
		defer configFile.Close()
		if err != nil {
			jww.ERROR.Println("Cannot create file .stor.yml in current directory.")
			os.Exit(1)
		}
		_, err = configFile.WriteString(skeleton_config_file)
		if err != nil {
			jww.ERROR.Println("Could not write to file .stor.yml.")
			os.Exit(1)
		}
		fmt.Println("Created hidden file .stor.yml in current working directory")
		fmt.Println("  Fill it out and then move it to your $HOME.")
		fmt.Println("  Generated .stor.yml documents settings required.")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
