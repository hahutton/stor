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

package blob

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	jww "github.com/spf13/jwalterweatherman"
)

type TransferType int

const MAX_BLOCKS = 50000 //This is the max per Azure Storage

const (
	IN TransferType = iota
	OUT
	BETWEEN
)

type Container struct {
	Alias       string
	AccountName string
	Name        string
	Endpoint    *url.URL
	Key         string
}

type Transfer struct {
	Container   *Container
	Type        TransferType
	FileName    string
	BlobName    string
	Parallelism int
	FileSize    int64
	BlockSize   int
}

type SigningRequest struct {
	Verb              string
	ContentEncoding   string
	ContentLanguage   string
	ContentLength     int
	ContentMD5        string
	ContentType       string
	Date              string
	IfModifiedSince   string
	IfMatch           string
	IfNoneMatch       string
	IfUnmodifiedSince string
	Range             string
	Account           string
	Container         string
	FileName          string
	BlockId           string
	TypeName          string
}

var client = retryablehttp.NewClient()

func init() {
	//Turn off debug???
	client.Logger = nil
}

func (c *Container) TransferIn(fileName string, fileSize int64, blobName string, blockSize int) *Transfer {
	return &Transfer{
		Container:   c,
		Type:        IN,
		FileName:    fileName,
		BlobName:    blobName[1:],  //TODO evaluate this. Done quickly
		Parallelism: 8,  //TODO used???
		FileSize:    fileSize,
		BlockSize:   blockSize,
	}
}

//for the defer which needs something to call
func returnToken(tokenBucket chan<- interface{}, token interface{}) {
	tokenBucket <- token
}

func makePrefix(fileName string) string {
	h := md5.New()
	io.WriteString(h, fileName)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func makeBlockId(prefix string, count int) string {
	id := fmt.Sprintf("%s%5d", prefix, count)
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func (c *Container) ListBlobs(prefix string) int {
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	tmpl, err := template.New("get_blob_list_auth_header").Parse(get_blob_list_auth_header)
	if err != nil {
		jww.ERROR.Println("bad tmpl")
		jww.ERROR.Println(err)
	}

	s := SigningRequest{"GET",
		"",
		"",
		0,
		"",
		"",
		datetime,
		"",
		"",
		"",
		"",
		"",
		c.AccountName,
		c.Name,
		"",
		"",
		"container",
	}

	var builder strings.Builder
	tmpl.Execute(&builder, s)

	decodedKey, err := base64.StdEncoding.DecodeString(c.Key)
	if err != nil {
		jww.ERROR.Println("Bad base64 key: ", err)
		os.Exit(1)
	}

	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(builder.String()))
	authKey := fmt.Sprintf("SharedKey %s:%s", c.AccountName, base64.StdEncoding.EncodeToString(h.Sum(nil)))

	target := fmt.Sprintf("%s?restype=container&comp=list", c.Endpoint)
	jww.TRACE.Println(target)

	req, err := retryablehttp.NewRequest("GET", target, nil)
	if err != nil {
		jww.ERROR.Println("Bad build of http request structure.", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", authKey)
	req.Header.Add("Date", datetime)
	req.Header.Add("x-ms-version", "2017-11-09")

	res, err := client.Do(req)
	if err != nil {
		jww.ERROR.Println("bad http request.", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	var results EnumerationResults

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		jww.ERROR.Println(err)
		os.Exit(1)
	}

	err = xml.Unmarshal(resBody, &results)
	if err != nil {
		jww.ERROR.Println(err)
		os.Exit(1)
	}

	for _, blob := range results.Blobs {
		fmt.Printf("%s %s %s %10d %s\n", blob.BlobType, blob.CreationTime, blob.Etag, blob.ContentLength, blob.Name)
	}
	return res.StatusCode
}

func (t *Transfer) putBlock(blockId string, b []byte, n int) int {
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	tmpl, err := template.New("put_block_auth_header").Parse(put_block_auth_header)
	if err != nil {
		fmt.Println("bad tmpl")
		fmt.Println(err)
	}

	s := SigningRequest{"PUT",
		"",
		"",
		n,
		"",
		"",
		datetime,
		"",
		"",
		"",
		"",
		"",
		t.Container.AccountName,
		t.Container.Name,
		t.BlobName,
		blockId,
		"block",
	}

	var builder strings.Builder
	tmpl.Execute(&builder, s)

	decodedKey, err := base64.StdEncoding.DecodeString(t.Container.Key)
	if err != nil {
		jww.ERROR.Println("Bad base64 key: ", err)
		os.Exit(1)
	}

	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(builder.String()))
	authKey := fmt.Sprintf("SharedKey %s:%s", t.Container.AccountName, base64.StdEncoding.EncodeToString(h.Sum(nil)))

	target := fmt.Sprintf("%s/%s?comp=block&blockid=%s", t.Container.Endpoint, t.BlobName, blockId)

	body := bytes.NewReader(b[:n])
	req, err := retryablehttp.NewRequest("PUT", target, body)
	if err != nil {
		jww.ERROR.Println("Bad build of http request structure.", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", authKey)
	req.Header.Add("Date", datetime)
	req.Header.Add("x-ms-version", "2017-11-09")

	res, err := client.Do(req)
	if err != nil {
		jww.ERROR.Println("bad http request.", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	resBody, _ := ioutil.ReadAll(res.Body)
	jww.TRACE.Println(res)
	jww.TRACE.Printf("%s\n", resBody)

	return res.StatusCode
}

func (t *Transfer) putBlockList(blockList []string) int {
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	bodyTemplate, err := template.New("put_block_list_body").Parse(put_block_list_body)
	if err != nil {
		jww.ERROR.Println("Bad put_block_list_body.tmpl", err)
		os.Exit(1)
	}
	var bodyBuilder strings.Builder
	bodyTemplate.Execute(&bodyBuilder, blockList)

	tmpl, err := template.New("put_block_list_auth_header").Parse(put_block_list_auth_header)
	if err != nil {
		jww.ERROR.Println("Bad put_block_list_auth_header.tmpl", err)
		os.Exit(1)
	}

	s := SigningRequest{"PUT",
		"",
		"",
		len(bodyBuilder.String()),
		"",
		"",
		datetime,
		"",
		"",
		"",
		"",
		"",
		t.Container.AccountName,
		t.Container.Name,
		t.BlobName,
		"",
		"blocklist",
	}

	var builder strings.Builder
	tmpl.Execute(&builder, s)

	decodedKey, err := base64.StdEncoding.DecodeString(t.Container.Key)
	if err != nil {
		jww.ERROR.Println("Bad base64 key: ", err)
		os.Exit(1)
	}

	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(builder.String()))
	authKey := fmt.Sprintf("SharedKey %s:%s", t.Container.AccountName, base64.StdEncoding.EncodeToString(h.Sum(nil)))

	target := fmt.Sprintf("%s/%s?comp=blocklist", t.Container.Endpoint, t.BlobName)

	body := strings.NewReader(bodyBuilder.String())
	req, err := retryablehttp.NewRequest("PUT", target, body)
	if err != nil {
		jww.ERROR.Println("Bad build of http request structure.", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", authKey)
	req.Header.Add("Date", datetime)
	req.Header.Add("x-ms-version", "2017-11-09")

	res, err := client.Do(req)
	if err != nil {
		jww.ERROR.Println("bad http request.", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	resBody, _ := ioutil.ReadAll(res.Body)
	jww.TRACE.Println(res)
	jww.TRACE.Printf("%s\n", resBody)

	return res.StatusCode
}

func (t *Transfer) Do() (statusCode int, err error) {

	var wg sync.WaitGroup

	//This is the max number of tokens in the bucket which limits concurrency
	bucketSize := runtime.NumCPU() * 10 //TODO parameterize this const
	tokenBucket := make(chan interface{}, bucketSize)
	for i := 0; i < bucketSize; i++ {
		tokenBucket <- i
	}

	//This is the number of work items which can be more or less than
	//the bucketSize which will determine whether this processing is gated
	blockCount := int(math.Ceil(float64(t.FileSize) / float64(t.BlockSize)))
	if blockCount > MAX_BLOCKS {
		jww.ERROR.Printf("Too many blocks. Max is %d. %d requested. Maybe adjust blockSize?", MAX_BLOCKS, blockCount)
		os.Exit(1)
	}
	blockList := make([]string, blockCount)

	blockIdPrefix := makePrefix(t.FileName)

	//Expect the number of go func calls to equal the number of blocks which are requests
	//Need to wait on all of them to finish
	wg.Add(blockCount)

	file, err := os.Open(t.FileName)
	defer file.Close()
	if err != nil {
		jww.ERROR.Println("Cannot open file for reading: ", t.FileName)
		jww.ERROR.Println(err)
		os.Exit(1)
	}

	buffer := make([]byte, t.BlockSize)

	for i := range blockList {
		token := <-tokenBucket
		n, err := file.Read(buffer)
		jww.TRACE.Printf("Read to block[%d] with length %d bytes", i, n)
		if err != nil {
			jww.ERROR.Println("Bad buffer read", err)
			os.Exit(1)
		}

		b := append([]byte{}, buffer[:n]...)
		go func(buf []byte, length int, count int, token interface{}) {
			defer returnToken(tokenBucket, token)
			defer wg.Done()

			blockId := makeBlockId(blockIdPrefix, count)
			jww.TRACE.Printf("Putting to blockList[%d] value %s", count, blockId)
			blockList[count] = blockId

			t.putBlock(blockId, buf, length)
		}(b, n, i, token)
	}
	wg.Wait()

	jww.TRACE.Println("BlockList to put:", blockList)
	resStatusCode := t.putBlockList(blockList)

	return resStatusCode, nil
}
