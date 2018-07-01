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

package providers

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

//These may drive to provider level config of blocks since max here is 10,000 * 100MB=1TB max blob size for stor tool
// or maybe config file general level of config where could assume file system as target or source
// and thereby up the max vals for a particular provider instead of lowest common denominator
const (
	MAX_BLOCKS     = 50000             //Azure Max block count
	MIN_BLOCK_SIZE = 1024 * 5
	MAX_BLOCK_SIZE = 1024 * 1024 * 100 //Azure max size
)

type BlobInfo struct {
	Name         string
	PathName     string
	CreatedAt    time.Time
	LastModified time.Time
	Length       int64
	Etag         string
	Encoding     string
	Type         string
	MD5          string
	BlobType     string
	IsDir        bool
}

type Block struct {
	Bytes   []byte
	Ordinal int
}

type Provider interface {
	Create(name string, stream <-chan *Block, blockCount int, tokenBucket chan int) error
	Open(name string, stream chan<- *Block, tokenBucket chan int, blockCount int, blockSize int) error
	Glob(pattern string) []*BlobInfo
	Stat(name string) *BlobInfo
	ProviderName() string
}

func Create(alias string) Provider {
	providerName := viper.GetString(fmt.Sprintf("aliases.%s.provider", alias))

	switch providerName {
	case "azure":
		sub := viper.Sub(fmt.Sprintf("aliases.%s", alias))
		return &AzureProvider{
			sub.GetString("accountName"),
			sub.GetString("name"),
			sub.GetString("key"),
		}
	case "file":
		return &FileProvider{}
	}
	jww.ERROR.Println("No provider found for alias: ", alias)
	jww.ERROR.Println("Check config or alias matching?")
	os.Exit(1)
	//Won't get here
	return nil
}

var aliasMatcher *regexp.Regexp = regexp.MustCompile("//([A-Za-z]+)(/.*)")

func Parse(aliasedPath string) (alias string, pathName string) {
	if !isAlias(aliasedPath) {
		return "file", aliasedPath
	}

	matches := aliasMatcher.FindStringSubmatch(aliasedPath)
	jww.TRACE.Println("//<alias>/<blobName> matches ->", matches)
	if len(matches) != 3 {
		jww.ERROR.Println("Bad alias. Won't parse.", aliasedPath)
		jww.ERROR.Println("Matches:", matches)
		os.Exit(1)
	}

	alias = matches[1]
	pathName = matches[2]
	return alias, pathName
}

func InitTokenBucket() chan int {
	viper.SetDefault("max_concurrency", runtime.NumCPU() * 10)
	bucketSize := viper.GetInt("max_concurrency")
	jww.TRACE.Println("Token Bucket Size: ", bucketSize)
	tokenBucket := make(chan int, bucketSize)
	for i := 0; i < bucketSize; i++ {
		tokenBucket <- i
	}
	return tokenBucket
}

func CalculateBlocks(info *BlobInfo) (int, int) {
	blockSize := viper.GetInt("blockSize")

	if blockSize < MIN_BLOCK_SIZE {
		jww.INFO.Println("Set Blocksize to minimum allowed:", MIN_BLOCK_SIZE)
		blockSize = MIN_BLOCK_SIZE
	}

	if blockSize > MAX_BLOCK_SIZE {
		jww.INFO.Println("Set Blocksize to maximum allowed:", MAX_BLOCK_SIZE)
		blockSize = MAX_BLOCK_SIZE
	}

	blockCount := int(math.Ceil(float64(info.Length) / float64(blockSize)))
	if blockCount > MAX_BLOCKS {
		jww.ERROR.Printf("Too many blocks. Max is %d. %d requested. Maybe adjust blockSize?", MAX_BLOCKS, blockCount)
		os.Exit(1)
	}
	return blockCount, blockSize
}

func isAlias(arg string) bool {
	return strings.HasPrefix(arg, "//")
}
