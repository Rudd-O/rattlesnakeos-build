# Build RattlesnakeOS manually

This process is certainly easier to get started with than [building via Jenkins](jenkins.md), but you don't get the automation and OTA updates that come with Jenkins builds.

## Prerequisites

Your build machine must be Debian 9, have at least 16 GB of RAM, and 200 GB of free disk space.  You will need Go 1.10 or later as well, present in your `$PATH`.

## Configure build parameters

If so desired, create a `custom-config.json` file.  This is a [JSON configuration file that allows you to control what goes into your images](customconfig.md).

## Check out source code

Place the file `render.go` from this project in a directory of your machine.

Now, in the same directory, `git clone` the RattlesnakeOS stack (https://github.com/dan-v/rattlesnakeos-stack) -- this will end up in a subdirectory `rattlesnakeos-stack`.

## Compile build script

From the abovementioned directory you'll run now:

```
GOPATH=$PWD/rattlesnakeos-stack go run render.go [...options...] -output stack-builder
```

The options are as follows:

*  `-build-type` string: build type (user or userdebug) (default `user`)
*  `-chromium-version` string: build with a specific version of Chromium
*  `-custom-config` string: path to a JSON file that has customizations (patches, script, prebuilts, et cetera) 
*  `-device` string: build the stack for this device (default "marlin")
*  `-hosts-file-url` string: build with a custom hosts file from an URL
*  `-ignore-version-checks`: ignore version checks altogether, building again
*  `-output` string: output file for stack script. (default "stack-builder")
*  `-release-download-address` string: URL where the Android platform will look for published updates

Of these, the ones most important are `-device` and `-build-type`.  Device refers to your device's code name, and build type lets you choose whether to do a `userdebug` build (debuggable but insecure) or a standard `user` build .

Once you've `go run` the program, you'll get a program `stack-builder` in the main directory.  This is your build script.

*Note:* as you can see, you can compile the build script on a separate machine that is not the build machine, then copy it to the build machine.

## Create main directory

On the build machine, create some directory where the entire build will happen.  In this example, it will be `/mnt/rattlesnakeos`.  Copy the build script `stack-builder` generated above to this directory.

## Deploy the keys

[After generating your device's signing keys](signingkeys.md), deploy them as follows.

Create directory `s3/rattlesnakeos-keys` under the main directory.

Place the generated keys in the `s3/rattlesnakeos-keys/<PRODUCT_NAME>` folder under the main directory.  Keep the directory structure.  In other words, the keys folder of your device will end up in `/mnt/rattlesnakeos/s3/rattlesnakeos-keys/<PRODUCT_NAME>`.

*Secure these keys*.

## Run build script

You're ready to go.  From the main directory, run `./stack-builder <your device name>` and the build will start.

## Manually flash the `*-factory-latest.tar.xz` once

The resulting images will be under `<main directory>/s3/rattlesnakeos-release/`.  You can find the factory latest tarball there.

Unpack the factory latest tarball.  Then flash the built image to your phone using the standard `fastboot` flashing procedure documented everywhere.  You'll find it in the artifacts page of the build (and, if you so chose, your release Web server as well).
