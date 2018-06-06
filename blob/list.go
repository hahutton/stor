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
