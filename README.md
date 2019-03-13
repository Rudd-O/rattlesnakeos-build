# Build RattlesnakeOS without the cloud

This project will help you build RattlesnakeOS (vanilla AOSP for Google Pixel and HiKey devices) directly on your hardware, so you need not trust "the cloud" with your phone's security.  This build is based on the RattlesnakeOS release schedule, complete with fully compliant secure boot and anti-theft protection.  It also completely RattlesnakeOS' Amazon AWS dependencies, running code only on machines you control fully.

You can use this project in one of two ways:

1. Run the build locally directly.  This build recipe will generate the necessary build script, which you can then run (after setting up signing keys).  You can then manually flash the result to your phone.
2. [Use Jenkins](jenkins.md).  This build recipe uses a Jenkinsfile and some custom code to adapt [the RattlesnakeOS build stack](https://github.com/dan-v/rattlesnakeos-stack/) for building Android directly on-prem.
  * This build recipe will also build periodically (by default, between the fifth and the fifteenth of each month, as per the `Jenkinsfile` triggers), as well as within every push to this repo (or your repo, if you fork this repo to your own).  This allows you to stay up-to-date with the latest security patches.  Of course, the build can manage an Android OTA update repo, so that updates hit your phone automatically.

Among the chief improvements over RattlesnakeOS is incremental build speed.  Failed or interrupted builds can be retried and will pick up exactly from where the failed build left off.  Source code is reused between builds as well.  Furthermore, if a successful build has taken place in the past, and nothing has changed from the previous build, the pipeline will exit early with a successful status.  You do not need to worry about wasting CPU, memory, disk space or bandwidth on repeat builds of the same thing.

## To-do

* Add support for Bromite.
