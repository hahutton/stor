![stor](https://github.com/hahutton/stor/raw/master/Docs/img/stor.logo.png)

Opinionated Azure Storage cmd line tool for copying, listing and remove blobs with NO DEPENDENCIES

## Installation

Single binary built specifically for macosx, linux and windows thanks to Go.


```bash
wget  https://hahutton.blob.core.windows.net/stor/downloads/macos/stor
```

```bash
wget  https://hahutton.blob.core.windows.net/stor/downloads/linux/stor
```

```bash
wget  https://hahutton.blob.core.windows.net/stor/downloads/windows/stor
```

### **stor** Root Command 

The **root** command provides top level usage and shows the full set of subcommands.
**stor** subcommands provide the functionality.

After **stor** has been downloaded, run it to see the various sub commands: 

```bash
➜  ~ ./stor 
stor is a cli tool to interact with azure storage.
While stor aims to be a sharp tool with a more unix philosophy,
azcopy should always be considered. stor aims to have no dependencies
which is a difference.

Usage:
  stor [command]

Available Commands:
  cp          copy source destination
  help        Help about any command
  ls          ls /alias/path...
  rm          remove a blob or blobs

Flags:
      --config string   config file (default is ./.stor.yml then $HOME/.stor.yml)
  -h, --help            help for stor
  -t, --trace           very verbose
  -v, --verbose         verbose

Use "stor [command] --help" for more information about a command.
```
 
### **stor** cp

The cp (copy) command pushes files to Azure Storage or pulls blobs to the local file system.
It performs either function in parallel by breaking individual files into blocks which are
PUT or requesting different ranges of blobs to GET with multiple, concurrent http calls to  
Azure Storage Restful APIs.


### **stor** ls

The ls (list) command lists the blobs in a container with globbing analagous to shells (without **).
The switches change the output and fmt of the results.

### **stor** rm

The rm (remove) command deletes blobs. Many blobs can be removed at once with pattern globbing.


### **stor** init

The init (initialize) command creates a simple .stor.yml file in the pwd (present working directory). It can be 
placed in the user's $HOME directory where it will be used as the default configuration for stor. It can be passed to
stor with a --config parameter to override that or even placed in the pwd of stor when executed (.).

## **stor** Core Features

**stor** performs all data movement commands concurrently by breaking files/blobs into blocks which can be moved
over https with multiple GETs or PUTs. The files/blobs are either chunked or stitched previous to PUTs or after GETs.
With sufficient bandwidth and memory moves can be maximized for available resources. Both parallelism and block size
can be configured. 

