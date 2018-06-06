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
	"regexp"
	"strings"

	"github.com/hahutton/stor/blob"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const AZ_STORAGE_BASE = "blob.core.windows.net"

func isGlob(path string) bool {
	return strings.Contains(path, "*")
}

func isDir(path string) bool {
	return strings.HasSuffix(path, "/")
}

func isAlias(arg string) bool {
	return strings.HasPrefix(arg, "//")
}

var aliasMatcher *regexp.Regexp = regexp.MustCompile("//([A-Za-z]+)/(.+)")

func parse(aliasedPath string) (container *blob.Container, pathName string, blockSize int) {
	matches := aliasMatcher.FindStringSubmatch(aliasedPath)
	jww.TRACE.Println("//<alias>/<blobName> matches ->", matches)
	if len(matches) != 3 {
		jww.ERROR.Println("Bad alias. Won't parse.", aliasedPath)
		jww.ERROR.Println("Matches:", matches)
		//TODO return an error instead of exiting here
		os.Exit(1)
	}

	alias := matches[1]
	pathName = matches[2]

	accountName := viper.GetString(fmt.Sprintf("aliases.%s.accountName", alias))
	containerName := viper.GetString(fmt.Sprintf("aliases.%s.name", alias))
	url, err := url.Parse(fmt.Sprintf("https://%s.%s/%s", accountName, AZ_STORAGE_BASE, containerName))
	if err != nil {
		jww.ERROR.Println("Bad names for account/container?", err)
		os.Exit(1)
	}

	container = &blob.Container{
		Alias:       alias,
		AccountName: accountName,
		Name:        containerName,
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
	blockSize = viper.GetInt("blockSize")

	return container, pathName, blockSize
}
