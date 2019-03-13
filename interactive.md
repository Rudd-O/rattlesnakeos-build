# Build RattlesnakeOS manually

This process is certainly easier to get started with than [building via Jenkins](jenkins.md), but you don't get the automation and OTA updates that come with Jenkins builds.

## Prerequisites

Your machine must be Debian 9, have at least 16 GB of RAM, and 200 GB of free disk space.  You will need Go 1.10 or later as well, present in your `$PATH`.

## Create main directory

Create some directory where the entire build will happen.  In this example, it will be `/mnt/rattlesnakeos`.

## Check out source code

Place the file `render.go` from this project in the main directory.

Now `git clone` the RattlesnakeOS stack (https://github.com/dan-v/rattlesnakeos-stack) -- this will end up in a subdirectory `rattlesnakeos-stack` of the main directory.

## Deploy the keys

[After generating your device's signing keys](signingkeys.md), deploy them as follows.

Create directory `s3/rattlesnakeos-keys` under the main directory.

Place the generated keys in the `s3/rattlesnakeos-keys/<PRODUCT_NAME>` folder under the main directory.  Keep the directory structure.  In other words, the keys folder of your device will end up in `/mnt/rattlesnakeos/s3/rattlesnakeos-keys/<PRODUCT_NAME>`.

*Secure these keys*.

## Configure build parameters

If so desired, create a `custom-config.json` file.  This is a [JSON configuration file that allows you to control what goes into your images](customconfig.md).

## Compile build script

From the main directory you'll now run:

```
GOPATH=$PWD/rattlesnakeos-stack go run render.go [...options...] -output stack-builder
```

The options are as follows:

*  -build-type string
  *  	build type (user or userdebug) (default "user")
*  -chromium-version string
  *  	build with a specific version of Chromium
*  -custom-config string
  *  	path to a JSON file that has customizations (patches, script, prebuilts, et cetera) 
*  -device string
  *  	build the stack for this device (default "marlin")
*  -hosts-file-url string
  *  	build with a custom hosts file from an URL
*  -ignore-version-checks
  *  	ignore version checks altogether, building again
*  -output string
  *  	Output file for stack script. (default "stack-builder")
*  -release-download-address string
  *  	URL where the Android platform will look for published updates

Of these, the ones most important are `-device` and `-build-type`.

Once you've `go run` the program, you'll get a program `stack-builder` in the main directory.

## Run build script

You're ready to go.  Run `./stack-builder <your device name>` and the build will start.
