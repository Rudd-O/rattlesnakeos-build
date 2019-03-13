# Release server (optional, `RELEASE_DOWNLOAD_ADDRESS` and `RELEASE_UPLOAD_ADDRESS`)

To support OTA updates [in your Jenkins build](jenkins.md), here's what you must do.

First, you gotta set up a Web server somewhere (and probably also add SSL certificates to it).  We will assume this server will be accessible at `https://yourserver.name/`, and the full URL to the OTA updates repo is `https://yourserver.name/ota-updates/`.  Instructions on how to set up a Web server are left to you.

## Web server on the same box as Jenkins master

If the server is on the same machine, ensure that the Jenkins master's UNIX username can write to the folder served at the URL `https://yourserver.name/ota-updates/`.

Test this part by hand.  Create a dummy Jenkins job that attempts to copy files into that folder.  Finally, see if the files are accessible through your Web browser.

## Remote Web server accessible via SSH

Ensure that your Jenkins master can SSH into the Web server and deploy files on the root directory of the Web server, such that the Jenkins master can publish the released files.  This will involve adding an UNIX user to the Web server machine, giving that user permission to write to the `ota-updates` subfolder served by the Web server, and setting up SSH pubkey authentication so that the Jenkins master can successfully `rsync` files into the `ota-updates` folder of the Web server (all via SSH).

Test this part by hand.  Create a dummy Jenkins job that attempts to push files from the Jenkins master via `rsync` or `scp`.  Finally, see if the files show when you browse to `https://yourserver.name/ota-updates/`, and the server allows you to download the pushed files.

## OTA updates release configuration

Now configure the mandatory defaults on your build pipeline.

Open the file `parameters.groovy` and adjust the `RELEASE_UPLOAD_ADDRESS` and `RELEASE_DOWNLOAD_ADDRESS` parameters:

For `RELEASE_UPLOAD_ADDRESS`, if you are using SSH to upload the files, it should be something like `remoteuser@remotehost:/path/to/webserver/ota-updates`.  If, however, the Web server is colocated with Jenkins, just specify an absolute path like `/srv/nginx/www-root/ota-updates`.

For `RELEASE_DOWNLOAD_ADDRESS`, in our example here, it should be something like `https://yourserver.name/ota-updates/`.

Have Jenkins rescan your multibranch job one more time so that the options are picked up.  Cancel the build that happens as a result of the rescan.
