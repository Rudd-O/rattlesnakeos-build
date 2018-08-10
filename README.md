# Build RattlesnakeOS in your Jenkins instance

This is a Jenkinsfile and additional code that will build RattlesnakeOS
(AOSP-based Android-like distribution) images for the Pixel XL, Pixel XL 2,
and other Google phones.  This build is based on the RattlesnakeOS release
schedule, complete with fully compliant secure boot and anti-theft protection.

This build recipe is based on [the RattlesnakeOS build stack](https://github.com/dan-v/rattlesnakeos-stack/).
This recipe mocks out the Amazon AWS aspects of the build process
and provides compatible replacements that will provide equivalent
functionality (with only minor effort on your part).

This build recipe will also build periodically.  If a successful build
has taken place in the past, the pipeline will exit early with a
successful status, so you do not need to worry about wasting CPU,
memory or disk space on repeat builds of the same thing.  The parameters
used to determine whether a build should run to completion are evident
from the pipeline script — check the script out if you want to know
what decides whether a build continues or not.

## How to use it

### Jenkins build slave configuration

Ensure you have a Debian 9-based Jenkins build slave configured and
working in your Jenkins master.

Ensure your build slave with 16 GB RAM and 200 GB disk space available.
Give that build slave the label `android`.

The `sudo` configuration needs to be adjusted in your build slave so that
the slave process can run commands as root via `sudo`.  The *Preparation*
stage of the build process will attempt to install several necessary packages
at the very beginning, by using `apt-get` with `sudo`.  This is bound to fail
on your system, unless you first install the packages in question.
In case of failure, run the build and see the log of the *Preparation*
stage -- then install the packages mentioned by the log.

### Signing keys generation (one-time-process)

Now note the device you'll use (e.g., `marlin`).  We'll use this shortly.

Create the keys now, using a secure machine that won't leak your keys.

#### Android Verified Boot 1.0 (`marlin` generation devices)

```
mkdir -p keys/marlin
cd keys/marlin
../../development/tools/make_key releasekey '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key platform '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key shared '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key media '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key verity '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
cd ../..
```

#### Android Verified Boot 2.0 (`taimen` generation devices)

```
mkdir -p keys/taimen
cd keys/taimen
../../development/tools/make_key releasekey '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key platform '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key shared '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key media '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
openssl genrsa -out avb.pem 2048
../../external/avb/avbtool extract_public_key --key avb.pem --output avb_pkmd.bin
cd ../..
```

### Jenkins master configuration (one-time process)

Add the shared Groovy library https://github.com/Rudd-O/shared-jenkins-libraries
to your Jenkins libraries configuration.

Create a Jenkins multibranch pipeline, pointed at this repository.

This will dispatch a build immediately.  Expect the build to fail, or cancel it
if you so desire.

In the Jenkins job folder (`$JENKINS_HOME/jobs/<your job name>`), create
a file named `parameters.groovy`, and supply default values in a Groovy
dictionary for which device you want to build (as in
`["DEVICE": "marlin"]`).

### Signing keys deployment (one-time process)

Place those keys in the `keys/<PRODUCT_NAME>` folder under the job directory
you created below the Jenkins `jobs` folder.  You must create one set of
keys per device.  *Secure these keys* because if you lose them, you won't be
able to create new flashable builds without unlocking and wiping your device.

Ensure the keys under this job directory are readable only by the Jenkins user.

### Release server (optional, one-time process)

Set up a Web server somewhere (and probably also add SSL certificates
to it).

Ensure that the Jenkins master can SSH into the Web server and deploy files
on the root directory of the Web server, such that the Jenkins master can
publish the released files.  This will involve adding an UNIX user, giving
it permission to write to the root of the Web server, and setting up
SSH pubkey authentication so that the Jenkins master can successfully `rsync`
files into the root of the Web server via SSH.

Test this part by hand.  Try pushing files via `rsync` as the Jenkins master,
and seeing if the Web server shows them and allows you to download them.

Now configure the mandatory defaults on your build pipeline:

In the Jenkins job folder (`$JENKINS_HOME/jobs/<your job name>`), create
a file named `parameters.groovy`, and configure two parameters in the
attendant Groovy dictionary:

1. `RELEASE_DOWNLOAD_ADDRESS`: this is your Web server URL that
   will show the published files to your phone (for the updater to work).
   Mind the slash at the end.
2. `RELEASE_UPLOAD_ADDRESS`: this is the address where the results
   will be published, in SSH address format
   `user@host:/path/to/root/of/webserver`.

### Rescan pipeline (one-time process)

Have Jenkins rescan your multibranch job so that the defaults are picked up.

### Finishing setup

You're now ready to go.

Verify that your defaults in `parameters.groovy`
got picked up by glancing at the *Build with parameters* page of your
build.

Build your first image.

Flash the built image to your phone using the standard `fastboot`
flashing procedure documented everywhere.  You'll find it
in the artifacts page of the build (and, if you so chose, your
release Web server).

## To-do

* Ensure that Chromium patched with bromite is stored with
  a different name (and the marker file as well) in "S3"
  than Chromium unpatched with bromite.  Otherwise one
  build will have a Chromium from a prior build that
  does not necessarily conform to the Chromium build
  parameters of this one build.
