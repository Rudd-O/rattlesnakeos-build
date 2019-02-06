def RELEASE_DOWNLOAD_ADDRESS = funcs.loadParameter('parameters.groovy', 'RELEASE_DOWNLOAD_ADDRESS', 'http://example.com/')
def RELEASE_UPLOAD_ADDRESS = funcs.loadParameter('parameters.groovy', 'RELEASE_UPLOAD_ADDRESS', '')
def CUSTOM_CONFIG = funcs.loadParameter('parameters.groovy', 'CUSTOM_CONFIG', '')
def HOSTS_FILE = funcs.loadParameter('parameters.groovy', 'HOSTS_FILE', '')

def ALL_DEVICES = ["marlin (Pixel XL)", "angler (Nexus 6P)", "bullhead (Nexus 5X)", "sailfish (Pixel)", "taimen (Pixel 2 XL)", "walleye (Pixel 2)", "hikey (HiKey)", "hikey960 (HiKey 960)"]
def DEVICE = funcs.loadParameter('parameters.groovy', 'DEVICE', "")
if (DEVICE != "") {
  DEVICE = [DEVICE] + ALL_DEVICES
} else {
  DEVICE = ALL_DEVICES
}

def runStack(currentBuild, actually_build, stage="") {
	def onlyReport = true
	def phase = "description"
	if (actually_build) {
		onlyReport = false
		if (stage == "") {
			phase = "build"
		} else {
			phase = stage
		}
	}
	def grepper = """#!/bin/bash -e
		grep '^aws_notify: ' android-build.log | sed 's/^aws_notify: //'
	"""
	script {
		try {
			sh """#!/bin/bash -e
			export HOME="\$PWD"
			export TMPDIR="\$PWD/tmp"
			mkdir -p "\$TMPDIR"
			export DEVICE=\$(echo "${params.DEVICE}" | cut -d ' ' -f 1)
			export STAGE=${stage}
			set -x
			set -o pipefail
			ONLY_REPORT=${onlyReport} ionice -c3 bash -x rattlesnakeos-stack/stack-builder "\$DEVICE" 2>&1 | tee android-build.log
			"""
			def description = sh (
				script: grepper,
				returnStdout: true
			).trim()
			if (description != "") {
				currentBuild.description = description
			}
		} catch(error) {
			currentBuild.description = "Failed in ${phase} phase:\n${error}.\n" + sh (
				script: grepper,
				returnStdout: true
			).trim() + "\n\n" + currentBuild.description
			throw error
		}
	}
}

// https://github.com/Rudd-O/shared-jenkins-libraries
@Library('shared-jenkins-libraries@master') _
pipeline {

	agent none

	triggers {
		pollSCM('* * * * *')
		cron('H 0,6,12,18 * * *')
	}

	options {
		disableConcurrentBuilds()
	}

	parameters {
		choice choices: DEVICE, description: 'The device model to build for.', name: 'DEVICE'
		choice choices: ["user", "userdebug"], description: 'The type of build you want.  Userdebug build types allow obtaining root via ADB, and enable ADB by default on boot.  See https://source.android.com/setup/build/building for more information.', name: 'BUILD_TYPE'
		string defaultValue: "", description: 'Version of Chromium to pin to if requested.', name: 'CHROMIUM_VERSION', trim: true
		string defaultValue: RELEASE_DOWNLOAD_ADDRESS, description: 'The HTTP(s) address, in http://host/path/to/folder/ format (note ending slash), where the published artifacts are exposed for the Updater app to download.  This is baked into your built release for the Updater app to use.  It is mandatory.', name: 'RELEASE_DOWNLOAD_ADDRESS', trim: true
		string defaultValue: RELEASE_UPLOAD_ADDRESS, description: 'The SSH address, in user@host:/path/to/folder format, to rsync artifacts to, in order to publish them.  Leave empty to skip publishing.', name: 'RELEASE_UPLOAD_ADDRESS', trim: true
		booleanParam defaultValue: false, description: 'Force build even if no new versions exist of components.', name: 'FORCE_BUILD'
		booleanParam defaultValue: false, description: 'Clean workspace completely before starting.  This will also force a build as a side effect.', name: 'CLEAN_WORKSPACE'
		text defaultValue: CUSTOM_CONFIG, description: 'An advanced option that allows you to specify customizations as explained in https://github.com/dan-v/rattlesnakos-stack/README.md .', name: 'CUSTOM_CONFIG'
		string defaultValue: HOSTS_FILE, description: 'An advanced option that allows you to specify a replacement /etc/hosts file to enable global dns adblocking (e.g. https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts).  Note: be careful with this, as you 1) will not get any sort of notification on blocking 2) if you need to unblock something you will have to rebuild the OS', name: 'HOSTS_FILE', trim: true
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
								relativeTargetDir: 'rattlesnakeos-stack'
							]],
							submoduleCfg: [],
							userRemoteConfigs: [[url: 'https://github.com/dan-v/rattlesnakeos-stack']]
						])
						script {
							dir("rattlesnakeos-stack") {
								env.RATTLESNAKEOS_GIT_HASH = sh (
									script: "git rev-parse --short HEAD",
									returnStdout: true
								).trim()
							}
							println "RattlesnakeOS Git hash is reported as ${env.RATTLESNAKEOS_GIT_HASH}"
							dir("rattlesnakeos-stack") {
								sh 'ln -sf . src'
							}
							dir("rattlesnakeos-stack/github.com/dan-v") {
								sh 'ln -sf ../../ rattlesnakeos-stack'
							}
							dir("rattlesnakeos-stack/stack") {
								writeFile name: "exports.go", content: '''package stack

func RenderTemplate(templateStr string, params interface{}) ([]byte, error) {
	return renderTemplate(templateStr, params)
}''')
							}
						}
					}
				}
				stage('Stash inputs') {
					steps {
						dir("../../../keys/") {
							stash includes: '**', name: 'keys'
						}
						stash includes: 'rattlesnakeos-stack/**', name: 'stack'
						dir("src") {
							stash includes: '**', name: 'code'
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
								sh "sudo rm -rf * .??*"
							}
						}
						stage("Unstash inputs") {
							steps {
								script {
									sh '''#!/bin/bash -xe
										rm -rf s3
										mkdir -p s3/rattlesnakeos-keys
									'''
								}
								dir("s3/rattlesnakeos-keys") {
									unstash 'keys'
								}
								sh 'rm -rf rattlesnakeos-stack'
								unstash 'stack'
								dir("rattlesnakeos-stack") {
									unstash 'code'
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
								script {
									try {
										copyArtifacts(
											projectName: JOB_NAME,
											selector: lastSuccessful(),
											excludes: '**/*tar.xz,**/*.zip'
										)
									} catch (hudson.AbortException e) {
										println "Artifacts from last build do not exist.  Continuing."
									}
								}
							}
						}
						stage("Deps") {
							steps {
								timeout(time: 10, unit: 'MINUTES') {
									retry(2) {
										script {
											funcs.aptInstall(["golang", "curl"])
										}
									}
								}
								script {
									funcs.aptEnableSrc()
								}
							}
						}
						stage("Stack") {
							steps {
								dir("rattlesnakeos-stack") {
									script {
										sh '''#!/bin/bash -ex
											env
		choice choices: DEVICE, description: 'The device model to build for.', name: 'DEVICE'
		choice choices: ["user", "userdebug"], description: 'The type of build you want.  Userdebug build types allow obtaining root via ADB, and enable ADB by default on boot.  See https://source.android.com/setup/build/building for more information.', name: 'BUILD_TYPE'
		string defaultValue: "", description: 'Version of Chromium to pin to if requested.', name: 'CHROMIUM_VERSION', trim: true
		string defaultValue: RELEASE_DOWNLOAD_ADDRESS, description: 'The HTTP(s) address, in http://host/path/to/folder/ format (note ending slash), where the published artifacts are exposed for the Updater app to download.  This is baked into your built release for the Updater app to use.  It is mandatory.', name: 'RELEASE_DOWNLOAD_ADDRESS', trim: true
		string defaultValue: RELEASE_UPLOAD_ADDRESS, description: 'The SSH address, in user@host:/path/to/folder format, to rsync artifacts to, in order to publish them.  Leave empty to skip publishing.', name: 'RELEASE_UPLOAD_ADDRESS', trim: true
		booleanParam defaultValue: false, description: 'Force build even if no new versions exist of components.', name: 'FORCE_BUILD'
		booleanParam defaultValue: false, description: 'Clean workspace completely before starting.  This will also force a build as a side effect.', name: 'CLEAN_WORKSPACE'
		custom_config
		string defaultValue: HOSTS_FILE, description: 'An advanced option that allows you to specify a replacement /etc/hosts file to enable global dns adblocking (e.g. https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts).  Note: be careful with this, as you 1) will not get any sort of notification on blocking 2) if you need to unblock something you will have to rebuild the OS', name: 'HOSTS_FILE', trim: true
											forcebuild=
											if [ "$FORCE_BUILD" == "true" ] ; then
												forcebuild=-force-build
											fi
											customconfig=
											if [ "$CUSTOM_CONFIG" != "" ] ; then
												echo "$CUSTOM_CONFIG" > custom-config.json
												customconfig="-custom-config custom-config.json"
											fi
											GOPATH="$PWD" go run ../render.go -output ../stack-builder \\
												-device "$DEVICE" \\
												-build-type "$BUILD_TYPE" \\
												-chromium-version "$CHROMIUM_VERSION" \\
												-release-upload-address "$RELEASE_DOWNLOAD_ADDRESS" \\
												$forcebuild \\
												$customconfig
										'''
									}
								}
							}
						}
						stage('Describe') {
							steps {
								timeout(time: 5, unit: 'MINUTES') {
									runStack(currentBuild, false)
								}
								script {
									if (currentBuild.description.contains("build not required")) {
										currentBuild.result = 'NOT_BUILT'
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
									runStack(currentBuild, true, "setup_env")
								}
							}
						}
						stage('check_chromium') {
							steps {
								timeout(time: 24, unit: 'HOURS') {
									runStack(currentBuild, true, "check_chromium")
								}
							}
						}
						stage('aosp_repo_init') {
							steps {
								timeout(time: 5, unit: 'MINUTES') {
									runStack(currentBuild, true, "aosp_repo_init")
								}
							}
						}
						stage('aosp_repo_modifications') {
							steps {
								timeout(time: 15, unit: 'MINUTES') {
									runStack(currentBuild, true, "aosp_repo_modifications")
								}
							}
						}
						stage('aosp_repo_sync') {
							steps {
								timeout(time: 3, unit: 'HOURS') {
									runStack(currentBuild, true, "aosp_repo_sync")
								}
							}
						}
						stage('aws_import_keys') {
							steps {
								timeout(time: 15, unit: 'MINUTES') {
									runStack(currentBuild, true, "aws_import_keys")
								}
							}
						}
						stage('setup_vendor') {
							steps {
								timeout(time: 1, unit: 'HOURS') {
									runStack(currentBuild, true, "setup_vendor")
								}
							}
						}
						stage('apply_patches') {
							steps {
								timeout(time: 30, unit: 'MINUTES') {
									runStack(currentBuild, true, "apply_patches")
								}
							}
						}
						stage('rebuild_marlin_kernel') {
							steps {
								timeout(time: 3, unit: 'HOURS') {
									runStack(currentBuild, true, "rebuild_marlin_kernel")
								}
							}
						}
						stage('build_aosp') {
							steps {
								timeout(time: 24, unit: 'HOURS') {
									runStack(currentBuild, true, "build_aosp")
								}
							}
						}
						stage('release') {
							steps {
								timeout(time: 15, unit: 'MINUTES') {
									runStack(currentBuild, true, "release")
								}
							}
						}
						stage('checkpoint_versions') {
							steps {
								timeout(time: 5, unit: 'MINUTES') {
									runStack(currentBuild, true, "checkpoint_versions")
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
				sh 'rm -rf s3'
				copyArtifacts(
					projectName: JOB_NAME,
					selector: specific(BUILD_NUMBER),
					filter: 's3/*-release/*ota_update*,s3/*-release/*-stable,s3/*-release/*-beta'
				)
				sh """
					rsync -a -- s3/*-release/ "${params.RELEASE_UPLOAD_ADDRESS}"/
				"""
				sh 'rm -rf s3'
				script {
					currentBuild.description = currentBuild.description + "\nPublished artifacts to ${params.RELEASE_UPLOAD_ADDRESS}"
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
