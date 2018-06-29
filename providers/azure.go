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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	jww "github.com/spf13/jwalterweatherman"
)

const AZ_STORAGE_BASE = "blob.core.windows.net"

type AzureProvider struct {
	AccountName   string
	ContainerName string
	Key           string
}

type signingRequest struct {
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

const get_blob_list_auth_header string = `{{ .Verb }}
{{ .ContentEncoding }}
{{ .ContentLanguage }}

{{ .ContentMD5 }}
{{ .ContentType }}
{{ .Date }}
{{ .IfModifiedSince }}
{{ .IfMatch }}
{{ .IfNoneMatch }}
{{ .IfUnmodifiedSince }}
{{ .Range }}
x-ms-version:2017-11-09
/{{ .Account }}/{{ .Container }}
comp:list
restype:container`

const put_block_auth_header string = `{{ .Verb }}
{{ .ContentEncoding }}
{{ .ContentLanguage }}
{{ .ContentLength }}
{{ .ContentMD5 }}
{{ .ContentType }}
{{ .Date }}
{{ .IfModifiedSince }}
{{ .IfMatch }}
{{ .IfNoneMatch }}
{{ .IfUnmodifiedSince }}
{{ .Range }}
x-ms-version:2017-11-09
/{{ .Account }}/{{ .Container }}/{{ .FileName }}
blockid:{{ .BlockId }}
comp:{{ .TypeName -}}`

const put_block_list_auth_header string = `{{ .Verb }}
{{ .ContentEncoding }}
{{ .ContentLanguage }}
{{ .ContentLength }}
{{ .ContentMD5 }}
{{ .ContentType }}
{{ .Date }}
{{ .IfModifiedSince }}
{{ .IfMatch }}
{{ .IfNoneMatch }}
{{ .IfUnmodifiedSince }}
{{ .Range }}
x-ms-version:2017-11-09
/{{ .Account }}/{{ .Container }}/{{ .FileName }}
comp:{{ .TypeName -}}`

const put_block_list_body string = `<?xml version="1.0" encoding="utf-8"?>
<BlockList>
{{range .}}
  <Uncommitted>{{.}}</Uncommitted>
{{end}}
</BlockList>`

type EnumerationResults struct {
	Blobs         []Blob `xml:"Blobs>Blob"`
	EndPoint      string `xml:"ServiceEndpoint,attr"`
	ContainerName string `xml:"ContainerName,attr"`
}

type Blob struct {
	Name               string `xml:"Name"`
	CreationTime       string `xml:"Properties>Creation-Time"`
	LastModified       string `xml:"Properties>Last-Modified"`
	Etag               string `xml:"Properties>Etag"`
	ContentLength      int64  `xml:"Properties>Content-Length"`
	ContentType        string `xml:"Properties>Content-Type"`
	ContentEncoding    string `xml:"Properties>Content-Encoding"`
	ContentLanguage    string `xml:"Properties>Content-Language"`
	ContentMD5         string `xml:"Properties>Content-MD5"`
	CacheControl       string `xml:"Properties>Cache-Control"`
	ContentDisposition string `xml:"Properties>Content-Disposition"`
	BlobType           string `xml:"Properties>BlobType"`
	AccessTier         string `xml:"Properties>AccessTier"`
	AccessTierInferred bool   `xml:"Properties>AccessTierInferred"`
	LeaseStatus        string `xml:"Properties>LeaseStatus"`
	LeaseState         string `xml:"Properties>LeaseState"`
	ServerEncrypted    bool   `xml:"Properties>ServerEncrypted"`
}

var client = retryablehttp.NewClient()

func init() {
	//Turn off debug???
	client.Logger = nil
}

func (azure *AzureProvider) returnToken(tokenBucket chan<- int, token int) {
	tokenBucket <- token
}

func (azure *AzureProvider) endPoint() string {
	return fmt.Sprintf("https://%s.%s/%s", azure.AccountName, AZ_STORAGE_BASE, azure.ContainerName)
}

func (azure *AzureProvider) putBlock(block *Block, blockId string, name string, token int) int {
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	tmpl, err := template.New("put_block_auth_header").Parse(put_block_auth_header)
	if err != nil {
		jww.ERROR.Println("bad tmpl")
		jww.ERROR.Println(err)
	}

	s := signingRequest{}
	s.Verb = "PUT"
	s.ContentLength = len(block.Bytes)
	s.Date = datetime
	s.Account = azure.AccountName
	s.Container = azure.ContainerName
	s.FileName = strings.TrimPrefix(name, "/")
	s.BlockId = blockId
	s.TypeName = "block"

	var builder strings.Builder
	tmpl.Execute(&builder, s)

	decodedKey, err := base64.StdEncoding.DecodeString(azure.Key)
	if err != nil {
		jww.ERROR.Println("Bad base64 key: ", err)
		os.Exit(1)
	}

	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(builder.String()))
	authKey := fmt.Sprintf("SharedKey %s:%s", azure.AccountName, base64.StdEncoding.EncodeToString(h.Sum(nil)))
	jww.TRACE.Println(builder.String())

	//Assumes name begins with '/'
	target := fmt.Sprintf("%s%s?comp=block&blockid=%s", azure.endPoint(), name, blockId)
	jww.TRACE.Println("target http request:", target)

	body := bytes.NewReader(block.Bytes)
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

func (azure *AzureProvider) putBlockList(name string, blockList []string) int {
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

	s := signingRequest{}
	s.Verb = "PUT"
	s.ContentLength = len(bodyBuilder.String())
	s.Date = datetime
	s.Account = azure.AccountName
	s.Container = azure.ContainerName
	s.FileName = strings.TrimPrefix(name, "/")
	s.TypeName = "blocklist"

	var builder strings.Builder
	tmpl.Execute(&builder, s)

	decodedKey, err := base64.StdEncoding.DecodeString(azure.Key)
	if err != nil {
		jww.ERROR.Println("Bad base64 key: ", err)
		os.Exit(1)
	}

	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(builder.String()))
	authKey := fmt.Sprintf("SharedKey %s:%s", azure.AccountName, base64.StdEncoding.EncodeToString(h.Sum(nil)))
	// Assumes name begins with '/'
	target := fmt.Sprintf("%s%s?comp=blocklist", azure.endPoint(), name)

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

func makeBlockId(prefix string, count int) string {
	id := fmt.Sprintf("%s%5d", prefix, count)
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func (azure *AzureProvider) Create(name string, stream <-chan *Block, blockCount int, tokenBucket chan int) error {
	//Init (AWS)
	//Put blocks -> fanout
	//  Waitgroup for these on
	//Put list/commit

	var wg sync.WaitGroup
	idList := make([]string, blockCount)

	for block := range stream {
		token := <-tokenBucket
		wg.Add(1)
		blockId := makeBlockId("stor", block.Ordinal)
		idList[block.Ordinal] = blockId

		go func(block *Block, blockId string, name string, token int) {
			defer azure.returnToken(tokenBucket, token)
			defer wg.Done()

			azure.putBlock(block, blockId, name, token)
		}(block, blockId, name, token)
		jww.INFO.Printf("Azure Provider Received Block[%d] with length %d", block.Ordinal, len(block.Bytes))
	}

	wg.Wait()

	azure.putBlockList(name, idList)

	return nil
}

func (azure *AzureProvider) Open(name string, stream chan<- *Block, tokenBucket chan int, blockCount int, blockSize int) error {
	return nil
}

func (azure *AzureProvider) Stat(name string) *BlobInfo {
	//There are no directories actually in blob stores
	blobInfo := &BlobInfo{}
	blobInfo.IsDir = false
	return blobInfo
}

func (azure *AzureProvider) Glob(pattern string) []*BlobInfo {
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	tmpl, err := template.New("get_blob_list_auth_header").Parse(get_blob_list_auth_header)
	if err != nil {
		jww.ERROR.Println("bad tmpl")
		jww.ERROR.Println(err)
	}

	s := signingRequest{}
	s.Verb = "GET"
	s.ContentLength = 0
	s.Date = datetime
	s.Account = azure.AccountName
	s.Container = azure.ContainerName
	s.TypeName = "container"

	var builder strings.Builder
	tmpl.Execute(&builder, s)

	decodedKey, err := base64.StdEncoding.DecodeString(azure.Key)
	if err != nil {
		jww.ERROR.Println("Bad base64 key: ", err)
		os.Exit(1)
	}

	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(builder.String()))
	authKey := fmt.Sprintf("SharedKey %s:%s", azure.AccountName, base64.StdEncoding.EncodeToString(h.Sum(nil)))

	target := fmt.Sprintf("%s?restype=container&comp=list", azure.endPoint())
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
	var layout string = "Mon, 02 Jan 2006 15:04:05 MST"
	matches := make([]*BlobInfo, len(results.Blobs))
	for i, blob := range results.Blobs {
		var blobInfo *BlobInfo = &BlobInfo{}
		blobInfo.Name = blob.Name
		blobInfo.PathName = blob.Name
		blobInfo.Length = blob.ContentLength
		blobInfo.CreatedAt, _ = time.Parse(layout, blob.CreationTime)
		blobInfo.LastModified, _ = time.Parse(layout, blob.LastModified)
		blobInfo.MD5 = blob.ContentMD5
		blobInfo.Etag = blob.Etag
		blobInfo.BlobType = blob.BlobType
		matches[i] = blobInfo
	}
	return matches
}
