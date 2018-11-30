vutils
======
A library of some high level utilities for the following:

- Load JSON Config files using vutils.Config
- Dealing with default values when using ENV variables using vutils.Defaults
- Launching of arbitrary commands sync/async with some useful functions using vutils.Exec
- Utilities to deal with files and directrories using vutils.Files
- A wrapper for vutils.Exec that deals with managing Processes: vutils.Procces
- A wrapper for golang.org/x/crypto/ssh for handling SSH sessions: vutils.SSH
- Some utilities for working with time.Time objects. At the moment it is just to establish if a time is in Daylight Savings or not: vutils.Time.IsDaylightSavingsTime(inTime time.Time) returns a true if it is.
- Some utilities for working with UUIDs (github.com/google/uuid): vutils.UUID

Installation
============

Installing with dep
-------------------
Run the following command in the root folder of your project. The vutils dependencies will be managed for you.
```
$ dep ensure -add github.com/768bit/vutils
```
You can install vutils using go get as below without the dependencies but you must do the below:
```
$ cd $GOPATH/src/github.com/768bit/vutils
$ dep ensure
```
Using go get
------------
Dependencies:
- github.com/google/uuid
- golang.org/x/crypto
- golang.org/x/sync

Install the above dependencies using go get then run the below command:
```
$ go get -u github.com/768bit/vutils
```
If you want to use dep to install the dependencies:
```
$ cd $GOPATH/src/github.com/768bit/vutils
$ dep ensure
```
Utilities
=========
Config
------
```
package main

import "github.com/768bit/vutils"
import "os"

func main() {
  cwd, _ := os.GetWd() //get the current working directory
  defList := []string{
    "./.config/configToLoad.json", //look in cwd
    "~/.config/configToLoad.json", //look in hole directory
    "/etc/folder/configToLoad.json", //look in /etc/folder
  }
  var config ConfigStruct

  //load the config from the available sources

  if err := vutils.Config.GetConfigFromDefaultList("ConfigID", cwd, defList, config); err != nil {
    panic(err)
  }

  //try save the config to the supplied source list

  if err, configPath := vutils.Config.TrySaveConfig(cwd, defList, config); err != nil {
    panic(err)
  } else {
    println("Config Saved to", configPath)
  }

  //directly save config to path

  if err, configPath := vutils.Config.SaveConfigToFile(cwd, "/etc/folder/configToLoad.json", config); err != nil {
    panic(err)
  } else {
    println("Config Saved to", configPath)
  }

}
```
Exec
----
See Exec.go for implementation
```
package main

import "github.com/768bit/vutils"
import "os"

func main() {
  //run ls -la, the false flag ensure that all output can be piped.
  acmd := vutils.Exec.CreateAsyncCommand("ls", false, "-la")
  acmd.SetWorkingDir("/tmp").CopyEnv().AddEnv("key", "value").BindToStdoutAndStdErr()
  //set working dir for command, cope the env variables from this process (parent),
  //add another env variable and bind the output to stdout and std err

  //binding of a SIGINT handler can be done using acmd.BindSigIntHandler()

  if err := acmd.StartAndWait(); err != nil {
    panic(err)
  }
}
```
License
=======
MIT Licensed. See LICENSE file.
