def RELEASE_DOWNLOAD_ADDRESS = funcs.loadParameter('parameters.groovy', 'RELEASE_DOWNLOAD_ADDRESS', 'http://example.com/')
def SKIP_CHROMIUM_BUILD = funcs.loadParameter('parameters.groovy', 'SKIP_CHROMIUM_BUILD', false)
def RELEASE_UPLOAD_ADDRESS = funcs.loadParameter('parameters.groovy', 'RELEASE_UPLOAD_ADDRESS', '')
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
			currentBuild.description = sh (
				script: grepper,
				returnStdout: true
			).trim()
		} catch(error) {
			currentBuild.description = "Failed in ${phase} phase:\n${error}.\n" + sh (
				script: grepper,
				returnStdout: true
			).trim()
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
		cron('H * * * *')
	}

	options {
		disableConcurrentBuilds()
	}

	parameters {
		choice choices: DEVICE, description: 'The device model to build for.', name: 'DEVICE'
		choice choices: ["user", "userdebug"], description: 'The type of build you want.  Userdebug build types allow obtaining root via ADB, and enable ADB by default on boot.  See https://source.android.com/setup/build/building for more information.', name: 'BUILD_TYPE'
		string defaultValue: RELEASE_DOWNLOAD_ADDRESS, description: 'The HTTP(s) address, in http://host/path/to/folder/ format (note ending slash), where the published artifacts are exposed for the Updater app to download.  This is baked into your built release for the Updater app to use.  It is mandatory.', name: 'RELEASE_DOWNLOAD_ADDRESS', trim: true
		string defaultValue: RELEASE_UPLOAD_ADDRESS, description: 'The SSH address, in user@host:/path/to/folder format, to rsync artifacts to, in order to publish them.  Leave empty to skip publishing.', name: 'RELEASE_UPLOAD_ADDRESS', trim: true
		booleanParam defaultValue: SKIP_CHROMIUM_BUILD, description: 'Skip Chromium build if a build already exists.', name: 'SKIP_CHROMIUM_BUILD'
		booleanParam defaultValue: false, description: 'Force build even if no new versions exist of components.', name: 'FORCE_BUILD'
		booleanParam defaultValue: false, description: 'Clean workspace completely before starting.  This will also force a build as a side effect.', name: 'CLEAN_WORKSPACE'
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
						}
					}
				}
				stage('Stash inputs') {
					steps {
						dir("../../../keys/") {
							stash includes: '**', name: 'keys'
						}
						sh '''
							set -ex
							mv rattlesnakeos-stack/templates/build_template.go .
							rm -f rattlesnakeos-stack/templates/*
							mv build_template.go rattlesnakeos-stack/templates/
						'''
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
									excludes: '**/*tar.xz,**/*.zip,**/*.apk'
								)
							} catch (hudson.AbortException e) {
								println "Artifacts from last build do not exist.  Continuing."
							}
						}
					}
				}
				stage("Deps") {
					steps {
						println "Install deps"
						timeout(time: 10, unit: 'MINUTES') {
							retry(2) {
								script {
									funcs.aptInstall(["golang", "curl"])
								}
							}
						}
						println "Enable source"
						script {
							funcs.aptEnableSrc()
						}
					}
				}
				stage("Stack") {
					steps {
						dir("rattlesnakeos-stack") {
							script {
								sh """#!/bin/bash -e
									go build main.go
									./main -output stack-builder \\
										-force-build="${params.FORCE_BUILD}" \\
										-skip-chromium-build="${params.SKIP_CHROMIUM_BUILD}" \\
										-release-url="${params.RELEASE_DOWNLOAD_ADDRESS}" \\
										-build-type="${params.BUILD_TYPE}"
									cat 'stack-builder' | nl -ha -ba -fa | sed 's/^/stack-builder: /'
								"""
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
				stage("Get old build") {
					when {
						expression {
							return currentBuild.result != 'NOT_BUILT' && !params.CLEAN_WORKSPACE
						}
					}
					steps {
						script {
							try {
								copyArtifacts(
									projectName: JOB_NAME,
									selector: lastSuccessful(),
								)
							} catch (hudson.AbortException e) {
								println "Artifacts from last build do not exist.  Continuing."
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
						stage('fetch_aosp_source') {
							steps {
								timeout(time: 6, unit: 'HOURS') {
									runStack(currentBuild, true, "fetch_aosp_source")
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
						stage('aws_import_keys') {
							steps {
								timeout(time: 1, unit: 'MINUTES') {
									runStack(currentBuild, true, "aws_import_keys")
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
						stage('aws_release') {
							steps {
								timeout(time: 15, unit: 'MINUTES') {
									runStack(currentBuild, true, "aws_release")
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
				stage('Stash artifacts') {
					when {
						expression {
							return currentBuild.result != 'NOT_BUILT'
						}
					}
					steps {
						stash includes: 's3/*-release/**', name: 'artifacts'
					}
				}
			}
		}
		stage('Finish') {
			agent { label 'master' }
			options { skipDefaultCheckout() }
			stages {
				stage('Unstash artifacts') {
					when {
						expression {
							return currentBuild.result != 'NOT_BUILT'
						}
					}
					steps {
						script {
							sh 'rm -rf s3'
						}
						unstash 'artifacts'
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
				stage('Publish') {
					when {
						expression {
							return currentBuild.result != 'NOT_BUILT' && params.RELEASE_UPLOAD_ADDRESS != ''
						}
					}
					steps {
						script {
							sh """#!/bin/bash -xe
								rsync -a -- s3/*-release/ "${params.RELEASE_UPLOAD_ADDRESS}"/
							"""
							currentBuild.description = currentBuild.description + "\nPublished artifacts to ${params.RELEASE_UPLOAD_ADDRESS}"
						}
					}
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
