![stor](https://github.com/hahutton/stor/raw/master/docs/img/stor.logo.png)

Opinionated Azure Storage cmd line tool for copying, listing and remove blobs with NO DEPENDENCIES

## Installation

Single binary built specifically for macosx, linux and windows thanks to Go.


```bash
wget  https://hahutton.blob.core.windows.net/stor/downloads/darwin_386-0.2.30.0.zip
wget  https://hahutton.blob.core.windows.net/stor/downloads/darwin_amd64-0.2.30.0.zip
```

```bash
wget  https://hahutton.blob.core.windows.net/stor/downloads/linux_386-0.2.30.0.zip
wget  https://hahutton.blob.core.windows.net/stor/downloads/linux_amd64-0.2.30.0.zip
```

```bash
wget  https://hahutton.blob.core.windows.net/stor/downloads/windows_386-0.2.30.0.zip
wget  https://hahutton.blob.core.windows.net/stor/downloads/windows_amd64-0.2.30.0.zip
```

### **stor** Command 

The **root** command provides top level usage and shows the full set of subcommands.
**stor** subcommands provide the functionality.

After **stor** has been downloaded, run it to see the various sub commands: 

```bash
➜  stor git:(master) ✗ stor
stor is a cli tool to interact with azure storage.
While stor aims to be a sharp tool with a more unix philosophy,
azcopy should be used whenever possible due to its robustness and
feature set. stor aims to have no dependencies which is a difference.

Usage:
  stor [command]

Available Commands:
  cp          Copy blobs between providers with cp like semantics
  help        Help about any command
  init        Create a skeleton config file
  ls          List blobs
  version     version information

Flags:
      --config string   config file (default is ./.stor.yml then $HOME/.stor.yml)
  -h, --help            help for stor
  -t, --trace           very verbose
  -v, --verbose         verbose

Use "stor [command] --help" for more information about a command.
```
 
### **stor** cp

The cp (copy) command pushes files to Azure Storage from the local file system.
It performs the function in parallel by breaking individual files into blocks which are
PUT with multiple, concurrent http calls to Azure Storage Restful APIs.

### **stor** ls

The ls (list) command lists the blobs in a container with prefix matching which is what most
cloud object stores provide. Beyond the appearance of a file system, cloud oject stores namespaces
are actually flat.

The switches change the output and fmt of the results.

### **stor** version

The version command outputs the binary's version.


### **stor** init

The init (initialize) command creates a simple .stor.yml file in the pwd (present working directory). It can be 
placed in the user's $HOME directory where it will be used as the default configuration for stor. It can be passed to
stor with a --config parameter to override that or even placed in the pwd of stor when executed (.).

## Core Features

**stor** performs all data movement commands concurrently by breaking files/blobs into blocks which can be moved
over https with multiple PUTs. The files/blobs are chunked previous to PUTs.
With sufficient bandwidth and memory moves can be maximized for available resources. Both parallelism and block size
can be configured. **stor** retries http reqeusts on failure but does not store any state.

**stor** breaks files/blobs into blocks which allow for parallel processing. The block size can be set in the .stor.yml
configuration file as can the max_concurrency setting. These parameters directly impact memory usage and the upper bounds
of a files size. 

