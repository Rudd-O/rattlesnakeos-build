# Build RattlesnakeOS with Jenkins

The process is admittedly more involved than the standard RattlesnakeOS stack, and it's harder to setup than [a manual build](interactive.md), but once it's setup, it's fire and forget.

The instructions require you to have a Jenkins master running on some physical machine, a sufficiently-powerful Jenkins slave to perform the builds on (perhaps the Jenkins master is powerful enough), and administrative privileges on both machines.

Follow the *Initial setup* instructions below, then just let your build server build Android for you.  Builds happen daily at 12 AM from the fifth to the fifteenth of the month.  You can also trigger manual builds (with tweaked settings) from the *Build with parameters* page of your build server, as often as you'd like.

OTA updates will be pushed (if you configured that) at the end of each build -- in that case, your phone will check for updates and download the OTA.

## Jenkins build slave configuration

Ensure you have a Debian 9-based Jenkins build slave configured and working in your Jenkins master.  Ensure your build slave has at least 16 GB RAM and 200 GB disk space available.  Give that build slave the label `android`.

The `sudo` configuration on the slave needs to be adjusted, so that the slave process can run commands as root via `sudo`.  The *Preparation* stage of the build process will attempt to install several necessary packages at the very beginning, by using `apt-get` with `sudo`.  This is bound to fail on your system, unless you first install the packages in question. In case of failure, run the build and see the log of the *Preparation* stage -- then install the packages mentioned by the log.

*If you only have one machine, but it is sufficiently powerful*, then ensure it's running Debian 9 and the Jenkins master.  In your Jenkins configuration, add the label `android` to the master node, so that the script can allocate builds to the machine.  In this case, you must reconfigure `sudo` on the master, as explained just in the prior paragraph.

## Add the Jenkins pipeline

Add the shared Groovy library https://github.com/Rudd-O/shared-jenkins-libraries to your Jenkins Global Pipeline libraries configuration, as per https://jenkins.io/doc/book/pipeline/shared-libraries/ .

Create a Jenkins multibranch pipeline, pointed at this repository (or your fork of it).

This will dispatch a build immediately.  Expect the build to fail, or cancel it if you so desire.

## Deploy the keys

[After generating your device's signing keys](signingkeys.md), deploy them as follows.

Locate the folder of the Jenkins multibranch pipeline you created.  It should contain the Jenkins `config.xml` file associated with the job.  Tip: vanilla Jenkins installs the folder under `/var/lib/jenkins/jobs`.

Place the generated keys in the `keys/<PRODUCT_NAME>` folder under the job folder.  Keep the directory structure.  In other words, if your Jenkins job folder is `/var/lib/jenkins/jobs/RattlesnakeOS` then the keys folder of your device will end up in `/var/lib/jenkins/jobs/RattlesnakeOS/keys/<PRODUCT_NAME>`.

*Secure these keys and your build server* (ensure the keys under this job directory are readable only by the Jenkins user).

## Configure build parameters

The default parameters this project uses are unlikely to suit you.  Fortunately, you can control them.

In your Jenkins master, locate the same folder you deployed the keys to.  It should have the `config.xml` file we mentioned before.  You will add, in that folder, a file named `parameters.groovy`.

This file must contain *only a single* Groovy map.  Upon first load of the `Jenkinsfile`, as well as subsequent rescans of the multibranch repo, the parameters are loaded into memory and serve as the defaults that will be picked during automatic builds.  This means that every time you change this file, you must do a scan of the multibranch repo from the project page's left sidebar.

Here's a sample file (options in the sample will be explained below):

```
[
    "DEVICE": "taimen",
    "BUILD_TYPE": "user",
    "HOSTS_FILE_URL": "http://myserver.athome.local/android-hosts.txt",
    "CUSTOM_CONFIG": '''
    {
        "custom-patches": [
            {
                "repo": "https://github.com/RattlesnakeOS/community_patches",
                "patches": [
                    "00001-global-internet-permission-toggle.patch",
                    "00002-global-sensors-permission-toggle.patch",
                    "00003-disable-menu-entries-in-recovery.patch",
                    "00004-increase-default-maximum-password-length.patch"
                ]
            }
        ]
    }
    ''',
    "RELEASE_DOWNLOAD_ADDRESS": "http://myserver.athome.local/",
    "RELEASE_UPLOAD_ADDRESS": "deployer@myserver.athome.local:/srv/copperhead/",
]
```

Here's a quick reference to the parameters that this file takes.

* `DEVICE`: mandatory; refers to the variant of the device you are building for (`marlin`, `taimen`...).
* `BUILD_TYPE`: optional; refers to whether you want a `user` (default) or `userdebug` (insecure but debuggable) build.
* `HOSTS_FILE_URL`: optional; refers to an URL that will be included as `/etc/hosts` in your device images, useful for permanent ad blocking of known bad / spam / adware domains
* `CUSTOM_CONFIG`: optional; [refers to a JSON configuration file that allows you to control what goes into your images](customconfig.md).
* `RELEASE_DOWNLOAD_ADDRESS`: optional; this is your Web server URL that will show the published files to your phone ([for the updater to work](releaseserver.md).
* `RELEASE_UPLOAD_ADDRESS`: optional; this is the address where the results [will be published](releaseserver.md).  See below for information.

For more information on how to use these options, follow the links above.

Once you have edited this file on your Jenkins master, have the Jenkins project *Scan Multibranch Pipeline Now*.  This causes the build to pick up the new defaults.  Cancel any build that happens as a result of the rescan.

## Do your first build

You're now ready to go.

Verify that your defaults in `parameters.groovy` got picked up by glancing at the *Build with parameters* page of your build.

Build your first image.  This will take anywhere from six to twelve hours.  Relax, it's okay.  If the build is interrupted, subsequent builds will pick up from where the previous ones left off.  This is, by the way, a huge feature that RattlesnakeOS does not have.

## Manually flash the `*-factory-latest.tar.xz` once

Flash the built image to your phone using the standard `fastboot` flashing procedure documented everywhere.  You'll find it in the artifacts page of the build (and, if you so chose, your release Web server as well).
