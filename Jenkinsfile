def RELEASE_DOWNLOAD_ADDRESS = funcs.loadParameter('parameters.groovy', 'RELEASE_DOWNLOAD_ADDRESS', 'http://example.com/')
def RELEASE_UPLOAD_ADDRESS = funcs.loadParameter('parameters.groovy', 'RELEASE_UPLOAD_ADDRESS', '')
def CUSTOM_CONFIG = funcs.loadParameter('parameters.groovy', 'CUSTOM_CONFIG', '')
def HOSTS_FILE_URL = funcs.loadParameter('parameters.groovy', 'HOSTS_FILE_URL', '')

def ALL_DEVICES = ["marlin (Pixel XL)", "angler (Nexus 6P)", "bullhead (Nexus 5X)", "sailfish (Pixel)", "taimen (Pixel 2 XL)", "walleye (Pixel 2)", "hikey (HiKey)", "hikey960 (HiKey 960)", "crosshatch (Pixel 3 XL)", "blueline (Pixel XL)"]
def DEVICE = funcs.loadParameter('parameters.groovy', 'DEVICE', "")
if (DEVICE != "") {
  DEVICE = [DEVICE] + ALL_DEVICES
} else {
  DEVICE = ALL_DEVICES
}
def ALL_BUILD_TYPES = ["user", "userdebug"]
def BUILD_TYPE = funcs.loadParameter('parameters.groovy', 'BUILD_TYPE', "user")
if (BUILD_TYPE != "") {
  BUILD_TYPE = [BUILD_TYPE] + ALL_BUILD_TYPES
} else {
  BUILD_TYPE = ALL_BUILD_TYPES
}

def depsInstall() {
	sh '''
	test -f /etc/apt/sources.list.d/backports.list || {
	echo "deb http://ftp.debian.org/debian stretch-backports main" | sudo tee /etc/apt/sources.list.d/backports.list
	sudo dpkg --add-architecture i386
	apt-get update
	}
	'''
	funcs.aptInstall([
		"golang-1.11",
		"curl",
		"fuseext2",
		"lib32z1",
		"lighttpd",
		"python-pexpect",
		"xvfb",
		"x11-utils",
		"bsdiff",
		"openjdk-8-jre",
		"openjdk-8-jdk",
		"binutils",
		"bison",
		"bzip2",
		"cdbs",
		"curl",
		"dbus-x11",
		"dpkg-dev",
		"elfutils",
		"devscripts",
		"fakeroot",
		"flex",
		"g++",
		"git-core",
		"git-svn",
		"gperf",
		"libappindicator3-dev",
		"libasound2-dev",
		"libatspi2.0-dev",
		"libbrlapi-dev",
		"libbz2-dev",
		"libcairo2-dev",
		"libcap-dev",
		"libc6-dev",
		"libcups2-dev",
		"libcurl4-gnutls-dev",
		"libdrm-dev",
		"libelf-dev",
		"libffi-dev",
		"libgbm-dev",
		"libglib2.0-dev",
		"libglu1-mesa-dev",
		"libgnome-keyring-dev",
		"libgtk-3-dev",
		"libkrb5-dev",
		"libnspr4-dev",
		"libnss3-dev",
		"libpam0g-dev",
		"libpci-dev",
		"libpulse-dev",
		"libsctp-dev",
		"libspeechd-dev",
		"libsqlite3-dev",
		"libssl-dev",
		"libudev-dev",
		"libwww-perl",
		"libxslt1-dev",
		"libxss-dev",
		"libxt-dev",
		"libxtst-dev",
		"locales",
		"openbox",
		"p7zip",
		"patch",
		"perl",
		"pkg-config",
		"python",
		"python-cherrypy3",
		"python-crypto",
		"python-dev",
		"python-numpy",
		"python-opencv",
		"python-openssl",
		"python-psutil",
		"python-yaml",
		"rpm",
		"ruby",
		"subversion",
		"uuid-dev",
		"wdiff",
		"x11-utils",
		"xcompmgr",
		"xz-utils",
		"zip",
		"libbluetooth-dev",
		"libxkbcommon-dev",
		"realpath",
		"libc6-i386",
		"lib32gcc1",
		"lib32stdc++6",
		"libpulse0",
		"libbz2-1.0",
		"libappindicator3-1",
		"libasound2",
		"libatk1.0-0",
		"libatspi2.0-0",
		"libc6",
		"libcairo2",
		"libcap2",
		"libcups2",
		"libexpat1",
		"libffi6",
		"libfontconfig1",
		"libfreetype6",
		"libglib2.0-0",
		"libgnome-keyring0",
		"libgtk-3-0",
		"libpam0g",
		"libpango1.0-0",
		"libpci3",
		"libpcre3",
		"libpixman-1-0",
		"libspeechd2",
		"libstdc++6",
		"libsqlite3-0",
		"libuuid1",
		"libwayland-egl1-mesa",
		"libx11-6",
		"libx11-xcb1",
		"libxau6",
		"libxcb1",
		"libxcomposite1",
		"libxcursor1",
		"libxdamage1",
		"libxdmcp6",
		"libxext6",
		"libxfixes3",
		"libxi6",
		"libxinerama1",
		"libxrandr2",
		"libxrender1",
		"libxtst6",
		"zlib1g",
		"binutils-arm-linux-gnueabihf",
		"binutils-aarch64-linux-gnu",
		"binutils-mipsel-linux-gnu",
		"binutils-mips64el-linux-gnuabi64",
	])
}

def runStack(currentBuild, actuallyBuild, stage="") {
	def onlyReport = true
	def phase = "description"
	if (actuallyBuild) {
		onlyReport = false
		if (stage == "") {
			phase = "build"
		} else {
			phase = stage
		}
	}
	def grepper = """#!/bin/bash -e
		grep -a '^aws_notify: ' android-build.log | sed 's/^aws_notify: //'
		grep -a '^custom_config: ' android-build.log | sed 's/^custom_config: //'
	"""
	try {
		sh """#!/bin/bash -e
		export HOME="\$PWD"
		export TMPDIR="\$PWD/tmp"
		mkdir -p "\$TMPDIR"
		export DEVICE=\$(echo "${params.DEVICE}" | cut -d ' ' -f 1)
		export STAGE=${stage}
		set -o pipefail
		ret=0
		# We disable build number to prevent unnecessary regeneration of code.
		# Jenkins manages its build number separately from the Android build.
		NINJA_STATUS="[%f/%t/%o/%e]	" JENKINS_BUILD_NUMBER=\$BUILD_NUMBER BUILD_NUMBER= BASH_TRACE=bash.trace ONLY_REPORT=${onlyReport} ionice -c3 bash stack-builder "\$DEVICE" 2>&1 | tee android-build.log | sed 's/^/${phase}:	/' || ret=\$?
		if [ ${onlyReport} == false -o \$ret != 0 ] ; then
			sed 's/^/${phase} trace:	/' bash.trace >&2
		fi
		exit \$ret
		"""
		if (!actuallyBuild) {
			def description = funcs.wrapPre(funcs.escapeXml(sh (
				script: grepper,
				returnStdout: true
			).trim()))
			currentBuild.description = currentBuild.description + description
		}
	} catch (org.jenkinsci.plugins.workflow.steps.FlowInterruptedException interruptEx) {
		throw interruptEx
	} catch (error) {
		currentBuild.description = "<p>Failed in ${phase} phase: ${error}.</p>" + currentBuild.description
		throw error
	}
	return sh(
		script: '''#!/bin/sh
		cat android-build.log''',
		returnStdout: true
	).trim()
}

// https://github.com/Rudd-O/shared-jenkins-libraries
@Library('shared-jenkins-libraries@master') _
pipeline {

	agent none

	triggers {
		pollSCM('H H * * *')
		cron('H 0 5,6,7,8,9,10,11,12,13,14,15 * *')
	}

	options {
		disableConcurrentBuilds()
	}

	parameters {
		choice choices: DEVICE, description: 'The device model to build for.', name: 'DEVICE'
		choice choices: BUILD_TYPE, description: 'The type of build you want.  Userdebug build types allow obtaining root via ADB, and enable ADB by default on boot.  See https://source.android.com/setup/build/building for more information.', name: 'BUILD_TYPE'
		string defaultValue: "", description: 'Version of Chromium to pin to if requested.', name: 'CHROMIUM_VERSION', trim: true
		string defaultValue: RELEASE_DOWNLOAD_ADDRESS, description: 'The HTTP(s) address, in http://host/path/to/folder/ format (note ending slash), where the published artifacts are exposed for the Updater app to download.  This is baked into your built release for the Updater app to use.  It is mandatory.', name: 'RELEASE_DOWNLOAD_ADDRESS', trim: true
		string defaultValue: RELEASE_UPLOAD_ADDRESS, description: 'The SSH address, in user@host:/path/to/folder format, to rsync artifacts to, in order to publish them.  Leave empty to skip publishing.', name: 'RELEASE_UPLOAD_ADDRESS', trim: true
		booleanParam defaultValue: false, description: 'Build (likely incrementally) even if no new versions exist of components.', name: 'IGNORE_VERSION_CHECKS'
		booleanParam defaultValue: false, description: 'Clean workspace completely before starting.  This will also force a build as a side effect.', name: 'CLEAN_WORKSPACE'
		text defaultValue: CUSTOM_CONFIG, description: 'An advanced option that allows you to specify customizations for your ROM (see the README.md file of this project).', name: 'CUSTOM_CONFIG'
		string defaultValue: HOSTS_FILE_URL, description: 'An advanced option that allows you to specify an URL containing a replacement /etc/hosts file to enable global dns adblocking (e.g. https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts ).  Note: be careful with this, as you 1) will not get any sort of notification on blocking 2) if you need to unblock something you will have to rebuild the OS', name: 'HOSTS_FILE_URL', trim: true
	}

	stages {
		stage('Begin') {
			agent { label 'master' }
			stages {
				stage('Setup') {
					steps {
						script {
							funcs.announceBeginning()
							funcs.durable()
						}
					}
				}
				stage('Clean master') {
					when {
						expression {
							return params.CLEAN_WORKSPACE
						}
					}
					steps {
						sh 'git clean -fxd'
					}
				}
				stage('Check') {
					steps {
						script {
							sh """#!/bin/bash -xe
								keydir="../../../keys/${params.DEVICE}/"
								test -d "\$keydir" || {
									echo "Error: this program requires keys for your device to be in \$keydir within the Jenkins master.  Aborting." >&2
								}
							"""
						}
					}
				}
				stage('Get stack') {
					steps {
						checkout([
							$class: 'GitSCM',
							branches: [[name: '*/9.0']],
							doGenerateSubmoduleConfigurations: false,
							extensions: [[
								$class: 'RelativeTargetDirectory',
								relativeTargetDir: 'upstream/rattlesnakeos-stack'
							]],
							submoduleCfg: [],
							userRemoteConfigs: [[url: 'https://github.com/dan-v/rattlesnakeos-stack']]
						])
						updateBuildNumberDisplayName()
					}
				}
				stage('Stash inputs') {
					steps {
						dir("../../../keys/") {
							stash includes: '**', name: 'keys'
						}
						dir("upstream/rattlesnakeos-stack") {
							stash includes: '**', name: 'stack'
						}
						stash includes: '*.go', name: 'code'
						script {
							try {
								stash includes: '*.json', name: 'config'
							} catch(e) {
								println "Cannot stash the JSON config file.  Assuming not present."
							}
						}
					}
				}
			}
		}
		stage('Run') {
			agent { label 'android' }
			options { skipDefaultCheckout() }
			stages {
				stage('Prepare') {
					stages {
						stage('Clean slave') {
							when {
								expression {
									return params.CLEAN_WORKSPACE
								}
							}
							steps {
								deleteDir()
							}
						}
						stage("Unstash inputs") {
							steps {
								dir("s3/rattlesnakeos-keys") {
									unstash 'keys'
									sh 'pwd && ls -la'
								}
								dir("upstream/rattlesnakeos-stack") {
									deleteDir()
									unstash 'stack'
									unstash 'code'
									script {
										try {
											unstash 'config'
										} catch(e) {
											println "Cannot unstash the JSON config file.  Assuming not present."
										}
									}
								}
							}
						}
						stage("Markers") {
							when {
								expression {
									return !params.CLEAN_WORKSPACE
								}
							}
							steps {
								timeout(time: 10, unit: 'MINUTES') {
									script {
										sh 'rm -rf s3/*-release'
										try {
											copyArtifacts(
												projectName: JOB_NAME,
												selector: lastSuccessful(),
												excludes: '**/*tar.xz,**/*.zip,**/*.apk'
											)
										} catch (org.jenkinsci.plugins.workflow.steps.FlowInterruptedException interruptEx) {
											throw interruptEx
										} catch (e) {
											println "Artifacts from last successful build does not exist (${e}).  Continuing."
										}
									}
								}
							}
						}
						stage("Deps") {
							steps {
								timeout(time: 10, unit: 'MINUTES') {
									retry(2) {
										script {
											depsInstall()
										}
										sh '''
										mountpoint /rw && sudo mount -o remount,noatime /rw || true
										mountpoint /home && sudo mount -o remount,noatime /home || true
										'''
									}
								}
								script {
									funcs.aptEnableSrc()
								}
							}
						}
						stage("Stack") {
							steps {
								script {
									sh '''#!/bin/bash -ex
										env
										pushd upstream/rattlesnakeos-stack
										ignoreversionchecks=
										if [ "$IGNORE_VERSION_CHECKS" == "true" ] ; then
											ignoreversionchecks=-ignore-version-checks
										fi
										hostsfileurl=
										if [ "$HOSTS_FILE_URL" != "" ] ; then
											hostsfileurl="-hosts-file-url $HOSTS_FILE_URL"
										fi
										if [ "$CUSTOM_CONFIG" != "" ] ; then
											echo "$CUSTOM_CONFIG" > custom-config.json
										fi
										customconfig=
										if [ -f custom-config.json ] ; then
											customconfig="-custom-config custom-config.json"
										fi
										set -x
										/usr/lib/go-1.11/bin/go run render.go -output ../../stack-builder \\
											-device "$DEVICE" \\
											-build-type "$BUILD_TYPE" \\
											-chromium-version "$CHROMIUM_VERSION" \\
											-release-download-address "$RELEASE_DOWNLOAD_ADDRESS" \\
											$ignoreversionchecks \\
											$hostsfileurl \\
											$customconfig
										popd
									'''
								}
							}
						}
						stage('Describe') {
							steps {
								timeout(time: 5, unit: 'MINUTES') {
									script {
										def buildLog = runStack(currentBuild, false)
										if (currentBuild.description.contains("build not required")) {
											currentBuild.result = 'NOT_BUILT'
										}
									}
								}
							}
						}
					}
				}
				stage('Build') {
					when {
						expression {
							return currentBuild.result != 'NOT_BUILT'
						}
					}
					stages {
						stage('setup_env') {
							steps {
								timeout(time: 1, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "setup_env")
									}
								}
							}
						}
						stage('check_chromium') {
							steps {
								script {
									def buildLog = ""
									env.SHOULD_BUILD_CHROMIUM = "yes"
									timeout(time: 10, unit: 'MINUTES') {
										buildLog = runStack(currentBuild, true, "check_chromium")
									}
									if (buildLog.contains("just copying s3 chromium artifact")) {
										println "Copying MonochromePublic.apk from last successful build."
										try {
											copyArtifacts(
												projectName: JOB_NAME,
												selector: lastSuccessful(),
												filter: '**/MonochromePublic.apk'
											)
											sh 'ls -l s3/rattlesnakeos-release/chromium'
											env.SHOULD_BUILD_CHROMIUM = "no"
										} catch (org.jenkinsci.plugins.workflow.steps.FlowInterruptedException interruptEx) {
											throw interruptEx
										} catch (e) {
											println "MonochromePublic.apk from last successful build does not exist (${e}).  Continuing."
										}
									}
								}
							}
						}
						stage('fetch_chromium') {
							when {
								expression {
									return env.SHOULD_BUILD_CHROMIUM == "yes"
								}
							}
							steps {
								timeout(time: 6, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "fetch_chromium")
									}
								}
							}
						}
						stage('build_chromium') {
							when {
								expression {
									return env.SHOULD_BUILD_CHROMIUM == "yes"
								}
							}
							steps {
								timeout(time: 12, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "build_chromium")
									}
								}
							}
						}
						stage('aosp_repo_init') {
							steps {
								timeout(time: 1, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "aosp_repo_init")
									}
								}
							}
						}
						stage('aosp_repo_modifications') {
							steps {
								timeout(time: 30, unit: 'MINUTES') {
									script {
										runStack(currentBuild, true, "aosp_repo_modifications")
									}
								}
							}
						}
						stage('aosp_repo_sync') {
							steps {
								timeout(time: 6, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "aosp_repo_sync")
									}
								}
							}
						}
						stage('aws_import_keys') {
							steps {
								timeout(time: 15, unit: 'MINUTES') {
									script {
										runStack(currentBuild, true, "aws_import_keys")
									}
								}
							}
						}
						stage('setup_vendor') {
							steps {
								timeout(time: 1, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "setup_vendor")
									}
								}
							}
						}
						stage('apply_patches') {
							steps {
								timeout(time: 30, unit: 'MINUTES') {
									script {
										runStack(currentBuild, true, "apply_patches")
									}
								}
							}
						}
						stage('rebuild_marlin_kernel') {
							steps {
								timeout(time: 3, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "rebuild_marlin_kernel")
									}
								}
							}
						}
						stage('build_aosp') {
							steps {
								timeout(time: 24, unit: 'HOURS') {
									script {
										runStack(currentBuild, true, "build_aosp")
									}
								}
							}
						}
						stage('release') {
							steps {
								timeout(time: 60, unit: 'MINUTES') {
									script {
										runStack(currentBuild, true, "release")
									}
								}
							}
						}
						stage('aws_upload') {
							steps {
								timeout(time: 30, unit: 'MINUTES') {
									script {
										runStack(currentBuild, true, "aws_upload")
									}
								}
							}
						}
						stage('checkpoint_versions') {
							steps {
								timeout(time: 5, unit: 'MINUTES') {
									script {
										runStack(currentBuild, true, "checkpoint_versions")
									}
								}
							}
						}
					}
				}
				stage('Archive') {
					when {
						expression {
							return currentBuild.result != 'NOT_BUILT'
						}
					}
					steps {
						archiveArtifacts artifacts: 's3/*-release/**', fingerprint: true
					}
				}
			}
		}
		stage('Publish') {
			agent { label 'master' }
			options { skipDefaultCheckout() }
			when {
				expression {
					return params.RELEASE_UPLOAD_ADDRESS != '' && currentBuild.result != 'NOT_BUILT'
				}
			}
			steps {
				dir('s3') {
					deleteDir()
				}
				copyArtifacts(
					projectName: JOB_NAME,
					selector: specific(BUILD_NUMBER),
					filter: 's3/*-release/*ota_update*,s3/*-release/*-factory-*,s3/*-release/*-stable,s3/*-release/*-beta'
				)
				sh """
					rsync -a -- s3/*-release/ "${params.RELEASE_UPLOAD_ADDRESS}"/
				"""
				dir('s3') {
					deleteDir()
				}
				script {
					currentBuild.description = currentBuild.description + "<p>Published artifacts to ${params.RELEASE_UPLOAD_ADDRESS}</p>"
				}
			}
		}
	}
	post {
		always {
			node('master') {
				script {
					funcs.announceEnd(currentBuild.currentResult)
				}
			}
		}
	}
}
