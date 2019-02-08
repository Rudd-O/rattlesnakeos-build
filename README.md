# Build RattlesnakeOS without the cloud

This project will help you build RattlesnakeOS (vanilla AOSP for Google Pixel and HiKey devices) directly on your hardware, so you need not trust "the cloud" with your phone's security.  This build is based on the RattlesnakeOS release schedule, complete with fully compliant secure boot and anti-theft protection.

This build recipe uses a Jenkinsfile and some custom code to adapt [the RattlesnakeOS build stack](https://github.com/dan-v/rattlesnakeos-stack/) for building Android directly on-prem.  This recipe bypasses RattlesnakeOS' Amazon AWS stages, running code only on machines you control fully.

This build recipe will also build periodically (by default, between the fifth and the tenth of each month, as per the `Jenkinsfile` triggers), as well as within every push to this repo (or your repo, if you fork this repo to your own).  This allows you to stay up-to-date with the latest security patches.  Of course, this recipe can manage an Android OTA update repo, so that updates hit your phone automatically.

Among the chief improvement over RattlesnakeOS is incremental build speed.  Failed or interrupted builds can be retried and will pick up exactly from where the failed build left off.  Source code is reused between builds.  Furthermore, if a successful build has taken place in the past, and nothing has changed from the previous build, the pipeline will exit early with a successful status.  You do not need to worry about wasting CPU, memory, disk space or bandwidth on repeat builds of the same thing.

The instructions require you to have a Jenkins master running on some physical machine, a sufficiently-powerful Jenkins slave to perform the builds on (perhaps the Jenkins master is powerful enough), and administrative privileges on both machines.

## How to use it

### Jenkins build slave configuration (one-time process)

Ensure you have a Debian 9-based Jenkins build slave configured and working in your Jenkins master.  Ensure your build slave has at least 16 GB RAM and 200 GB disk space available.  Give that build slave the label `android`.

The `sudo` configuration on the slave needs to be adjusted, so that the slave process can run commands as root via `sudo`.  The *Preparation* stage of the build process will attempt to install several necessary packages at the very beginning, by using `apt-get` with `sudo`.  This is bound to fail on your system, unless you first install the packages in question. In case of failure, run the build and see the log of the *Preparation* stage -- then install the packages mentioned by the log.

*If you only have one machine, but it is sufficiently powerful*, then ensure it's running Debian 9 and the Jenkins master.  In your Jenkins configuration, add the label `android` to the master node, so that the script can allocate builds to the machine.  In this case, you must reconfigure `sudo` on the master, as explained just in the prior paragraph.

### Signing keys generation (one-time-process)

Now note the device you'll build images for (e.g., `marlin`).  We'll use this shortly.

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

Add the shared Groovy library https://github.com/Rudd-O/shared-jenkins-libraries to your Jenkins Global Pipeline libraries configuration, as per https://jenkins.io/doc/book/pipeline/shared-libraries/ .

Create a Jenkins multibranch pipeline, pointed at this repository (or your fork of it).

This will dispatch a build immediately.  Expect the build to fail, or cancel it if you so desire.

### Signing keys deployment (one-time process)

Locate the folder of the Jenkins multibranch pipeline you created.  It should contain the Jenkins `config.xml` file associated with the job.  Tip: vanilla Jenkins installs the folder under `/var/lib/jenkins/jobs`.

Place the generated keys in the `keys/<PRODUCT_NAME>` folder under the job folder.

*Secure these keys and your build server* (ensure the keys under this job directory are readable only by the Jenkins user).  If you lose the keys, you won't be able to create new flashable builds without unlocking and wiping your device.  If your keys are stolen, someone could upload a malicious ROM to your device without you noticing.

### Default build parameters (one time process)

The default parameters this project uses are unlikely to suit you.  Fortunately, you can control them.

In your Jenkins master, locate the same folder you deployed the keys to.  It should have the `config.xml` file we mentioned before.  You will add, in that folder, a file named `parameters.groovy`.

This file must contain *only a single* Groovy map (e.g. `["DEVICE": "taimen", "BUILD_TYPE": "userdebug"]`).  Upon first load of the `Jenkinsfile`, as well as subsequent rescans of the multibranch repo, the parameters are loaded into memory and serve as the defaults that will be picked during automatic builds.

The parameters that go in the map are documented fully in the *Build with parameters* page of the build project, but for completeness, I will briefly describe the most important here:

* `def DEVICE` refers to the variant of the device you are building for (`marlin`, `taimen`...).
* `def BUILD_TYPE` refers to whether you want a `user` (default) or `userdebug` (insecure but debuggable) build.
* `def CUSTOM_CONFIG` refers to a JSON configuration file that allows you to control what goes into your images (explained below).
* `def HOSTS_FILE_URL` refers to an URL that will be included as `/etc/hosts` in your device images, useful for permanent ad blocking of known bad / spam / adware domains

Once you have edited this file on your Jenkins master, have the Jenkins project *Scan Multibranch Pipeline Now*.  This causes the build to pick up the new defaults.  Cancel any build that happens as a result of the rescan.

#### RattlesnakeOS ROM customization options

RattlesnakeOS has a variety of customization options you can use in order to customize your build.  However, there is a caveat: unlike RattlesnakeOS, the configuration options aren't supplied in the same way to the build.

In RattlesnakeOS, you create a TOML configuration file `.rattlesnakeos.toml` using the `rattlesnakeos-stack config` command, then you edit and save the configuration file.  This configuration file is then used by `rattlesnakeos-stack deploy`.

In this project, you can do one of three things:

1. *Check in* the customization options as a JSON file named `custom-config.json`, alongside the `Jenkinsfile` within.  Yes, this requires you to fork the project.
2. If you'd like not to fork the project, you can also place the JSON text as the (string) value of the `CUSTOM_CONFIG` parameter of your `parameters.groovy` file.  This is obviously more complicated.
3. Finally, you can manually paste JSON text directly into the *Build with parameters* page.  Any text pasted there will override any existing `custom-config.json` file for that specific build.

JSON being not TOML, there are only syntax differences between what may go into the configuration file.

Let's use the RattlesnakeOS README example for our purpose here.  Suppose you'd put this into a `.rattlesnakeos.toml` config:

```
[[custom-patches]]
  repo = "https://github.com/RattlesnakeOS/community_patches"
  patches = [
      "00001-global-internet-permission-toggle.patch", "00002-global-sensors-permission-toggle.patch",
  ]

[[custom-scripts]]
  repo = "https://github.com/RattlesnakeOS/example_patch_shellscript"
  scripts = [ "00002-custom-boot-animation.sh" ]
```

Here's how the *exact same thing* would look like, in your `custom-config.json` file:

```
{
    "custom-patches": {
        "repo": "https://github.com/RattlesnakeOS/community_patches",
        "patches": [
            "00001-global-internet-permission-toggle.patch",
            "00002-global-sensors-permission-toggle.patch"
        ]
    },
    "custom-scripts": {
        "repo": "https://github.com/RattlesnakeOS/example_patch_shellscript",
        "scripts": [
            "00002-custom-boot-animation.sh"
        ]
    }
}
```

See?  Nothing extraordinary.  One small caveat: this program supports only a limited subset of options of the `.rattlesnakeos.toml` configuration file.  Here is a comprehensive list:

```
		CustomPatches:          customizations.CustomPatches,
		CustomScripts:          customizations.CustomScripts,
		CustomPrebuilts:        customizations.CustomPrebuilts,
		CustomManifestRemotes:  customizations.CustomManifestRemotes,
		CustomManifestProjects: customizations.CustomManifestProjects,
```

*For the programming-curious:* The data structures populated by both `.rattlesnakeos.toml` and `custom-config.json` are defined in file https://github.com/dan-v/rattlesnakeos-stack/blob/9.0/stack/aws.go . A full reference to the ROM customization options is available under the [the Customizations section of the RattlesnakeOS README](https://github.com/dan-v/rattlesnakeos-stack).

If you opted for the JSON-in-`parameters.groovy` option, have the Jenkins project *Scan Multibranch Pipeline Now*.  This causes the build to pick up the new defaults.  Cancel any build that happens as a result of the rescan, and manually dispatch one more build.  If you opted for the fork-and-check-in-my-own-`config.json` option, all you have to do is commit and push your changes — your build server will start to build.

### Release server (optional, one-time process)

To support OTA updates, here's what you must do.

First, you gotta set up a Web server somewhere (and probably also add SSL certificates to it).  We will assume this server will be accessible at `https://yourserver.name/`, and the full URL to the OTA updates repo is `https://yourserver.name/ota-updates/`.  Instructions on how to set up a Web server are left to you.

#### Web server on the same box as Jenkins master

If the server is on the same machine, ensure that the Jenkins master's UNIX username can write to the folder served at the URL `https://yourserver.name/ota-updates/`.

Test this part by hand.  Create a dummy Jenkins job that attempts to copy files into that folder.  Finally, see if the files are accessible through your Web browser.

#### Remote Web server accessible via SSH

Ensure that your Jenkins master can SSH into the Web server and deploy files on the root directory of the Web server, such that the Jenkins master can publish the released files.  This will involve adding an UNIX user to the Web server machine, giving that user permission to write to the `ota-updates` subfolder served by the Web server, and setting up SSH pubkey authentication so that the Jenkins master can successfully `rsync` files into the `ota-updates` folder of the Web server (all via SSH).

Test this part by hand.  Create a dummy Jenkins job that attempts to push files from the Jenkins master via `rsync` or `scp`.  Finally, see if the files show when you browse to `https://yourserver.name/ota-updates/`, and the server allows you to download the pushed files.

#### OTA updates release configuration

Now configure the mandatory defaults on your build pipeline.

Open the file `parameters.groovy` and add two keys to the map:

1. `RELEASE_DOWNLOAD_ADDRESS`: this is your Web server URL that will show the published files to your phone (for the updater to work).  In our example, it was `https://yourserver.name/ota-updates/`.
2. `RELEASE_UPLOAD_ADDRESS`: this is the address where the results will be published.
  * If you are using SSH to upload the files, it should be something like `remoteuser@remotehost:/path/to/webserver/ota-updates`.
  * If, however, the Web server is colocated with Jenkins, just specify an absolute path like `/srv/nginx/www-root/ota-updates`.

### Rescan pipeline (one-time process)

Have Jenkins rescan your multibranch job one more time so that the defaults are picked up.  Cancel the build that happens as a result of the rescan.

### Begin building (as often as you'd like)

You're now ready to go.

Verify that your defaults in `parameters.groovy` got picked up by glancing at the *Build with parameters* page of your build.

Build your first image.  This will take anywhere from six to twelve hours.  Relax, it's okay.  If the build is interrupted, subsequent builds will pick up from where the previous ones left off.  This is, by the way, a huge feature that RattlesnakeOS does not have.

Flash the built image to your phone using the standard `fastboot` flashing procedure documented everywhere.  You'll find it in the artifacts page of the build (and, if you so chose, your release Web server).

You are good to go now.  Enjoy!

## To-do

* Add support for Bromite.
