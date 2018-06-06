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
