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
	"os"
	"path/filepath"

	jww "github.com/spf13/jwalterweatherman"
)

type FileProvider struct {
}

func (fp *FileProvider) Create(name string, stream <-chan *Block, blockCount int, tokenBucket chan int) error {
	return nil
}

func (fp *FileProvider) Open(name string, stream chan<- *Block, tokenBucket chan int, blockCount int, blockSize int) error {
	jww.INFO.Printf("Open local file: %s with %d blocks of %d size", name, blockCount, blockSize)

	go func() {
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			jww.ERROR.Println("Bad local file open:", name)
			jww.ERROR.Println(err)
			os.Exit(1)
		}

		readBuffer := make([]byte, blockSize)

		for i := 0; i < blockCount; i++ {
			n, err := file.Read(readBuffer)
			if err != nil {
				jww.ERROR.Println("Bad block read from file: ", name)
				jww.ERROR.Println(err)
				os.Exit(1)
			}

			block := &Block{
				append([]byte{}, readBuffer[:n]...),
				i,
			}

			jww.INFO.Printf("Read to Block.Id[%d] with length %d bytes", block.Ordinal, len(block.Bytes))
			stream <- block
		}
		close(stream)
	}()

	return nil
}

func (fp *FileProvider) Stat(name string) *BlobInfo {
	fileInfo, err := os.Lstat(name)
	if err != nil {
		jww.ERROR.Println("Bad filename Stat:", name)
		jww.ERROR.Println(err)
		os.Exit(1)
	}
	var blobInfo *BlobInfo = &BlobInfo{}
	blobInfo.Name = fileInfo.Name()
	blobInfo.PathName = name
	blobInfo.Length = fileInfo.Size()
	blobInfo.LastModified = fileInfo.ModTime()
	blobInfo.BlobType = "FileSystem"
	blobInfo.IsDir = fileInfo.IsDir()
	return blobInfo
}

func (fp *FileProvider) Glob(pattern string) []*BlobInfo {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		jww.ERROR.Println("Bad filepath Glob:", pattern)
		jww.ERROR.Println(err)
		os.Exit(1)
	}

	matches := make([]*BlobInfo, len(paths))
	for i, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			jww.ERROR.Println("Bad filepath Stat:", path)
			jww.ERROR.Println(err)
			os.Exit(1)
		}
		var blobInfo *BlobInfo = &BlobInfo{}
		blobInfo.Name = fileInfo.Name()
		blobInfo.Length = fileInfo.Size()
		blobInfo.LastModified = fileInfo.ModTime()
		blobInfo.BlobType = "FileSystem"
		blobInfo.PathName = path
		matches[i] = blobInfo
	}
	return matches
}
