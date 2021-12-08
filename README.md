# revgen

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![ReportCard](https://goreportcard.com/badge/github.com/ejoffe/spr)](https://goreportcard.com/report/github.com/ejoffe/spr)
[![Doc](https://godoc.org/github.com/ejoffe/spr?status.svg)](https://godoc.org/github.com/ejoffe/spr)

## only run go:generate when code changes

At Inigo we generate a lot of go code. While our compile and build time runs in a matter of seconds thanks to the great go toolchain, running all the generators takes a few minutes on a good day. Running a single generator is fairly fast, but having to keep track of which generator to run gets annoying very fast.  
Revgen keeps track of all the go:generate dependencies and only runs the generators for code that has been updated. By running one to a handful of generators at most, the go:generate run time goes down from minutes to seconds.  
Each go:generate command is configured with a list of dependent files. When revgen is run, it calculates the hash of all these files, compares it with the latest hash, if they differ, runs the corresponding go:generate command and updates the stored hash.  

Installation
------------

### Go
```shell
> go install github.com/inigolabs/revgen@latest
```

### Brew
```shell
> brew tap inigolabs/homebrew-tap
> brew install revgen
```

### Manual
Download the right pre-compiled binary for your system from the [releases page](https://github.com/inigolabs/revgen/releases) and install.

Operation
---------
```shell
> revgen init
```
The very first time, run **revgen init** to initialize and create a **.revgen.yml** config file.  
The config file is placed in the root go directory.  
See the configuration section for more information on setting go generate file dependecies.
***  
```shell
> revgen
```
Run **revgen** anywhere inside your go workspace to call all the generators for code that has changed.  
***  
```shell
> revgen -force
```
Run **revgen -force** to run all the genrators regardless of code updates. 
***
```shell
> revgen update
```
Run **revgen update** to update the config file when go::generate commands are added or removed from the code. 
***
```shell
> revgen check
```
Run **revgen check** to check that all hashes match the current state of the code.  
Can be useful in continious integration pipelines to make sure all needed code has been generated.  
  
Configuration
-------------
> TODO
  
License
-------
- [MIT License](LICENSE)
  
Happy Coding!
-------------