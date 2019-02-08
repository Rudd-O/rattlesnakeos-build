package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"stack"
	"strings"
	"templates"
)

func replace(text string, original string, substitution string, numReplacements int) (string, error) {
	newText := strings.Replace(text, original, substitution, numReplacements)
	if text == newText {
		return "", fmt.Errorf("The replacement of\n%s\n\nfor\n%s\nproduced no changes", original, substitution)
	}
	return newText, nil
}

func alterTemplate(txt string) (string, error) {
	var replacements = []struct {
		original        string
		substitution    string
		numReplacements int
	}{
		{
			`#!/bin/bash`,
			`#!/bin/bash

if [ -n "$BASH_TRACE" ] ; then
	exec 19> "$BASH_TRACE"
	BASH_XTRACEFD=19
fi
set -x
`,
			-1,
		},
		{
			`"https://${AWS_RELEASE_BUCKET}.s3.amazonaws.com"`,
			`"<% .ReleaseDownloadAddress %>"`,
			-1,
		},
		{
			`AWS_SNS_ARN=$(aws --region ${REGION} sns list-topics --query 'Topics[0].TopicArn' --output text | cut -d":" -f1,2,3,4,5)":${STACK_NAME}"`,
			`AWS_SNS_ARN=none`,
			-1,
		},
		{`$(curl -s http://169.254.169.254/latest/meta-data/instance-type)`, "none", -1},
		{`$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | awk -F\" '/region/ {print $4}')`, "none", -1},
		{`$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)`, "none", -1},
		{
			`message="No build is required, but FORCE_BUILD=true"
      echo "$message"
`, `aws_notify "No build is required, but FORCE_BUILD=true"
`, -1,
		},
		{
			`message="No build is required, but IGNORE_VERSION_CHECKS=true"
      echo "$message"
`, `aws_notify "No build is required, but IGNORE_VERSION_CHECKS=true"
`, -1,
		},
		{`echo "New build is required"`, `aws_notify "New build is required"`, -1},
		{
			`BUILD_TYPE="user"`,
			`BUILD_TYPE="<% .BuildType %>"`,
			-1,
		},
		{
			"Stack Name: %s\\n  Stack Version: %s %s\\n  Stack Region: %s\\n  ", "", -1},
		{
			`"${STACK_NAME}" "${STACK_VERSION}" "${STACK_UPDATE_MESSAGE}" "${REGION}" `,
			"",
			-1,
		},
		{"Instance Type: %s\\n  Instance Region: %s\\n  Instance IP: %s\\n  ", "", -1},
		{`"${INSTANCE_TYPE}" "${INSTANCE_REGION}" "${INSTANCE_IP}" `, "", -1},
		{
			`repo init --manifest-url "$MANIFEST_URL" --manifest-branch "$AOSP_BRANCH" --depth 1 || true`,
			`repo init --manifest-url "$MANIFEST_URL" --manifest-branch "$AOSP_BRANCH" --depth 1 || true
  gitcleansources`,
			-1,
		},
		{
			`# copy to build tree
  mkdir -p ${BUILD_DIR}/external/chromium/prebuilt/arm64`,
			`# do not copy to build tree - later stage does it on demand
  # we just copy it to S3 so that the later stage can obtain it
`,
			-1,
		},
		{
			`cp out/Default/apks/MonochromePublic.apk ${BUILD_DIR}/external/chromium/prebuilt/arm64/`,
			`aws s3 cp out/Default/apks/MonochromePublic.apk "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk"`,
			-1,
		},
		{
			`  # upload to s3 for future builds
  aws s3 cp "${BUILD_DIR}/external/chromium/prebuilt/arm64/MonochromePublic.apk" "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk"`,
			`  # Suppressed copy to S3 as that has happened already - now we just save the built revision to S3`,
			-1,
		},
		{
			`aws s3 cp "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk" ${BUILD_DIR}/external/chromium/prebuilt/arm64/`,
			`# Suppressed copy from S3 to external/prebuilt/arm64/ as this happens later`,
			-1,
		},
		{`fetch --nohooks android`, `test -f .gclient || fetch --nohooks android`, -1},
		{
			`yes | gclient sync --with_branch_heads --jobs 32 -RDf

  # cleanup any files in tree not part of this revision
  git clean -dff

  # reset any modifications
  git checkout -- .`,
			`gitcleansources
  yes | gclient sync --with_branch_heads --jobs 32 -RDf`,
			-1,
		},
		{`gn gen out/Default`, `gitrestoretimestamps`, -1},
		{
			`patch_launcher
}`,
			`patch_launcher

  # Now we restore the timestamps we saved in gitcleansources.
  # If the files have actually not changed since the last build,
  # even after being patched or embedded from untarred tarballs,
  # then this will cause the incremental builds to go much faster.
  # since a few files that the patch process changes, don't actually
  # change between incremental builds -- only their timestamp
  # changes, but those files are at the very bottom of the tree
  # of dependencies, so they cause ninja to do a lot of thinking
  # and rebuilding.
  pushd ${BUILD_DIR}
  gitrestoretimestamps
  popd
}`,
			-1,
		},
		{`out/Default`, `"$HOME"/chromium-out`, -1},
		{`rm -rf $HOME/chromium`, `# We skip rm -rf'ing Chromium to avoid redownloading sources.`, -1},
		{
			"linux-image-$(uname --kernel-release)",
			"$(apt-cache search linux-image-* | awk ' { print $1 } ' | sort | egrep -v -- '(-dbg|-rt|-pae)' | grep ^linux-image-[0-9][.] | tail -1)",
			-1,
		},
		{
			`retry git clone`,
			`retry gitavoidreclone`,
			-1,
		},
		{
			`MARLIN_KERNEL_SOURCE_DIR="${HOME}/kernel/google/marlin"`,
			`MARLIN_KERNEL_SOURCE_DIR="${HOME}/kernel/google/marlin"
MARLIN_KERNEL_OUT_DIR="$HOME/kernel-out/$DEVICE"`,
			-1,
		},
		{
			`bash -c "\
    set -e;
    cd ${BUILD_DIR};
    . build/envsetup.sh;
    make -j$(nproc --all) dtc mkdtimg;
    export PATH=${BUILD_DIR}/out/host/linux-x86/bin:${PATH};
    ln --verbose --symbolic ${KEYS_DIR}/${DEVICE}/verity_user.der.x509 ${MARLIN_KERNEL_SOURCE_DIR}/verity_user.der.x509;
    cd ${MARLIN_KERNEL_SOURCE_DIR};
    make -j$(nproc --all) ARCH=arm64 marlin_defconfig;
    make -j$(nproc --all) ARCH=arm64 CONFIG_COMPAT_VDSO=n CROSS_COMPILE=${BUILD_DIR}/prebuilts/gcc/linux-x86/aarch64/aarch64-linux-android-4.9/bin/aarch64-linux-android-;
    cp -f arch/arm64/boot/Image.lz4-dtb ${BUILD_DIR}/device/google/marlin-kernel/;
    rm -rf ${BUILD_DIR}/out/build_*;
  "`,
			`bash -c "\
    set -e;
    mkdir -p ${MARLIN_KERNEL_OUT_DIR} ;
    cd ${BUILD_DIR};
    . build/envsetup.sh;
    set -x
    make -j$(nproc --all) dtc mkdtimg;
    export PATH=${BUILD_DIR}/out/host/linux-x86/bin:${PATH};
    ln --verbose --symbolic -f ${KEYS_DIR}/${DEVICE}/verity_user.der.x509 ${MARLIN_KERNEL_SOURCE_DIR}/verity_user.der.x509;
    cd ${MARLIN_KERNEL_SOURCE_DIR} ;
    make -j$(nproc --all) ARCH=arm64 marlin_defconfig O=${MARLIN_KERNEL_OUT_DIR} || make -j$(nproc --all) ARCH=arm64 mrproper marlin_defconfig O=${MARLIN_KERNEL_OUT_DIR} ;
    make -j$(nproc --all) ARCH=arm64 CONFIG_COMPAT_VDSO=n CROSS_COMPILE=${BUILD_DIR}/prebuilts/gcc/linux-x86/aarch64/aarch64-linux-android-4.9/bin/aarch64-linux-android- O=${MARLIN_KERNEL_OUT_DIR} ;
    # Now copy the recently-built kernel from its kernel-out place.
    rsync -a --inplace ${MARLIN_KERNEL_OUT_DIR}/arch/arm64/boot/Image.lz4-dtb ${BUILD_DIR}/device/google/marlin-kernel/Image.lz4-dtb;
    rm -rf ${BUILD_DIR}/out/build_*;
  "`,
			-1,
		},
		{
			`# make modifications to default AOSP`,
			`# make modifications to default AOSP
  # Since we just git cleaned everything, we will have to re-copy
  # the MonochromePublic.apk file from S3.
  mkdir -p ${BUILD_DIR}/external/chromium/prebuilt/arm64
  aws s3 cp "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk" ${BUILD_DIR}/external/chromium/prebuilt/arm64/
  `,
			-1,
		},
		{
			`timeout 30m "${BUILD_DIR}/vendor/android-prepare-vendor/execute-all.sh" --debugfs --keep --yes --device "${DEVICE}" --buildID "${AOSP_BUILD}" --output "${BUILD_DIR}/vendor/android-prepare-vendor"`,
			`mkdir -p "${HOME}/vendor-in"
  local flag="${HOME}/vendor-in/.${DEVICE}-$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")"
  if test -f "${flag}" ; then
    true
  else
    timeout 30m "${BUILD_DIR}/vendor/android-prepare-vendor/execute-all.sh" --fuse-ext2 --yes --device "${DEVICE}" --buildID "${AOSP_BUILD}" --output "${HOME}/vendor-in"
    touch "${flag}"
  fi`,
			-1,
		},
		{
			`mkdir --parents "${BUILD_DIR}/vendor/google_devices" || true
  rm -rf "${BUILD_DIR}/vendor/google_devices/$DEVICE" || true
  mv "${BUILD_DIR}/vendor/android-prepare-vendor/${DEVICE}/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/${DEVICE}" "${BUILD_DIR}/vendor/google_devices"

  # smaller devices need big brother vendor files
  if [ "$DEVICE" != "$DEVICE_FAMILY" ]; then
    rm -rf "${BUILD_DIR}/vendor/google_devices/$DEVICE_FAMILY" || true
    mv "${BUILD_DIR}/vendor/android-prepare-vendor/$DEVICE/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/$DEVICE_FAMILY" "${BUILD_DIR}/vendor/google_devices"
  fi`,
			`mkdir --parents "${BUILD_DIR}/vendor/google_devices"
  # Instead of destroying source files with mv (and then causing a lengthy rebuild due to execute-all.sh)
  # we mash the files into their final destination using rsync.  This also works additively since below
  # we can mash big brother devices' files using the same technique.
  # This saves an enormous amount of time.
  rsync -avHAX --inplace --delete --delete-excluded "${HOME}/vendor-in/${DEVICE}/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/${DEVICE}/" "${BUILD_DIR}/vendor/google_devices/${DEVICE}/"

  # smaller devices need big brother vendor files
  if [ "$DEVICE" != '$DEVICE_FAMILY' ]; then
    rsync -avHAX --inplace --delete --delete-excluded "${HOME}/vendor-in/$DEVICE/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/$DEVICE_FAMILY/" "${BUILD_DIR}/vendor/google_devices/$DEVICE_FAMILY/"
  fi`,
			-1,
		},
		{
			`source build/envsetup.sh`,
			`set +x ; source build/envsetup.sh ; set -x`,
			-1,
		},
		{
			`"$(wget -O - "${RELEASE_URL}/${RELEASE_CHANNEL}")"`,
			`"$(aws s3 cp "s3://${AWS_RELEASE_BUCKET}/${RELEASE_CHANNEL}" -)"`,
			-1,
		},
		{
			`make clobber`,
			`# do not make clobber, verity key generation happens only once`,
			-1,
		},
	}

	for _, r := range replacements {
		var err error
		if txt, err = replace(txt, r.original, r.substitution, r.numReplacements); err != nil {
			return "", err
		}
	}

	txt = strings.TrimSuffix(txt, "full_run\n")
	txt = txt + `# Beginning of outright overridden functions

aws() {
  quiet _aws "$@"
}

_aws() {
  func="$1"
  cmd="$2"
  in="$3"
  out="$4"
  if [ "$func" == "sns" ]
  then
	if [[ $7 == --message=* ]]
	then
		echo "${7#--message=}" | sed 's/^/aws_notify: /' >&2
		echo "$(dumpcustomconfig)" | sed 's/^/custom_config: /' >&2
	else
		echo "$8" | sed 's/^/aws_notify: /' >&2
		echo "$(dumpcustomconfig)" | sed 's/^/custom_config: /' >&2
	fi
  elif [ "$func" == "s3" ]
  then
	if [ "$cmd" == "cp" ]
	then
		if [[ $in == s3://* ]]
		then
			in="${in#s3://}"
			in="$HOME/s3/$in"
		fi
		if [[ $out == s3://* ]]
		then
			out="${out#s3://}"
			out="$HOME/s3/$out"
		fi
		if [ "$in" == "-" ]
		then
			mkdir -p $( dirname "$out" )
			cat - > "$out"
		elif [ "$out" == "-" ]
		then
			cat "$in"
		else
			mkdir -p $( dirname "$out" )
			cp -f --preserve=all "$in" "$out"
		fi
	elif [ "$cmd" == "ls" ]
	then
		if [[ $in == s3://* ]]
		then
			in="${in#s3://}"
			in="$HOME/s3/$in"
		fi
		ls -1 "$in"
	elif [ "$cmd" == "rm" ]
	then
		rm -f -- "$in"
	elif [ "$cmd" == "sync" ]
	then
		if [[ $in == s3://* ]]
		then
			in="${in#s3://}"
			in="$HOME/s3/$in"
		fi
		if [[ $out == s3://* ]]
		then
			out="${out#s3://}"
			out="$HOME/s3/$out"
		fi
		rsync -a --delete -- "$in/" "$out/"
	fi
  fi
}

gitavoidreclone() {
	if test -d "$2"/.git ; then
		pushd "$2"
		sed -i 's|url = .*|url = '"$1"'|' .git/config
		git fetch
		popd
	else
		git clone "$1" "$2"
	fi
}

quiet() {
	local r=0
	local cmd="$1"
	shift
	set +x
	"$cmd" "$@" || r="$?"
	set -x
	return "$r"
}

giterate() {
	local ret=0
	local cmd="$1"
	shift
	for gitdir in $(find -name .git -type d) ; do
		pushd "$gitdir/.." > /dev/null || continue
		"$cmd" "$@" || ret=$?
		popd > /dev/null
		if [ "$ret" != "0" ] ; then return "$ret" ; fi
	done
}

dumpcustomconfig() {
  pushd ${BUILD_DIR}
  if [ -f custom-config.json ] ; then
      echo "Custom configuration:"
      cat custom-config.json | sed 's/^/  /'
  else
      echo "No custom configuration."
  fi
  popd
}

gitcleansource() {
	local type
	local filename
	local sum
	local timestamp
	rm -f .git/timestampsums
	while read type filename ; do
                if [ -f "$filename" ] ; then
                        sum=$(md5sum "$filename" | awk ' { print $1 } ')
                        timestamp=$(stat -c %y "$filename")
                elif [ -e "$filename" ] ; then
                        sum="notafilenomd5sum"
                        timestamp=$(stat -c %y "$filename")
                else
                        sum="deletednomd5sum"
                        timestamp="no time stamp"
                fi
                echo "$sum $timestamp $filename" >> .git/timestampsums
	done < <(git status --ignored --porcelain)
	if [ -f .git/timestampsums ] ; then
		echo "$PWD: has modifications" >&2
		cat .git/timestampsums >&2
		git clean -fxd || return $?
		git reset --hard || return $?
	else
                echo "$PWD: does not have modifications" >&2
        fi
}

gitrestoretimestamp() {
	local sum
	local date
	local time
	local timezone
	local filename
	local timestamp
	local actualsum
	local actualtimestamp
	test -f .git/timestampsums || {
                echo "time restore $PWD: nothing to restore" >&2
                return 0
        }
	while read sum date time timezone filename ; do
		# If the file does not exist or is not a file,
		# don't bother restoring the timestamp
		test -f "$filename" || {
                    echo "time restore $PWD: $filename does not exist" >&2
                    continue
                }
		timestamp="$date $time $timezone"
		actualsum=$(md5sum "$filename" | awk ' { print $1 } ')
		actualtimestamp=$(stat -c %y "$filename")
		# If the file has changed, it has earned its mtime.
		if [ "$sum" != "$actualsum" ] ; then
                    echo "time restore $PWD: $filename sum differs" >&2
                    continue
                fi
		# If the file says "no time stamp", the file did not exist,
		# cannot restore timestamp.
		if [ "$timestamp" == "no time stamp" ] ; then
                    echo "time restore $PWD: $filename has no timestamp" >&2
                    continue
                fi
		# If the file has the same time stamp as before,
		# don't bother restoring it.
		if [ "$timestamp" == "$actualtimestamp" ] ; then
                    echo "time restore $PWD: $filename has kept its timestamp" >&2
                    continue
                fi
		echo "time restore $PWD: $filename unchanged, mtime differs, restoring" >&2
		touch -d "$timestamp" "$filename"
	done < <(cat .git/timestampsums)
}

gitcleansources() {
	giterate gitcleansource "$@"
}

gitrestoretimestamps() {
	giterate gitrestoretimestamp "$@"
}

aws_logging()
{
	return
}

cleanup() {
  rv=$?
  if [ $rv -ne 0 ]
  then
    aws_notify "RattlesnakeOS Build FAILED"
  fi
  exit $rv
}

persist_latest_versions() {
  rm -rf env*.save
  mkdir -p s3/interstage
  cat > s3/interstage/env.$BUILD_NUMBER.save <<EOF
STACK_UPDATE_MESSAGE="$STACK_UPDATE_MESSAGE"
LATEST_STACK_VERSION="$LATEST_STACK_VERSION"
LATEST_CHROMIUM="$LATEST_CHROMIUM"
FDROID_CLIENT_VERSION="$FDROID_CLIENT_VERSION"
FDROID_PRIV_EXT_VERSION="$FDROID_PRIV_EXT_VERSION"
AOSP_BUILD="$AOSP_BUILD"
AOSP_BRANCH="$AOSP_BRANCH"
EOF
}

reload_latest_versions() {
  source s3/interstage/env.$BUILD_NUMBER.save
}

get_encryption_key() {
  echo "Assert not reached ${FUNCNAME}." >&2 ; exit 100
}

initial_key_setup() {
  echo "Assert not reached ${FUNCNAME}." >&2 ; exit 100
}

gen_keys() {
  log_header "${FUNCNAME} (overridden)"

  pushd ${BUILD_DIR}
  gitrestoretimestamps
  popd

  if [ "${DEVICE}" == "marlin" ] || [ "${DEVICE}" == "sailfish" ]; then
    gen_verity_key "${DEVICE}"
  fi

  if [ "${DEVICE}" == "walleye" ] || [ "${DEVICE}" == "taimen" ]; then
    gen_avb_key "${DEVICE}"
  fi
}

aws_import_keys() {
  log_header "${FUNCNAME} (overridden)"
  aws s3 sync "s3://${AWS_KEYS_BUCKET}" "${KEYS_DIR}"
  gen_keys
}


if [ "$ONLY_REPORT" == "true" ]
then
full_run() {
  log_header ${FUNCNAME}

  get_latest_versions
  persist_latest_versions
  check_for_new_versions
}
else
full_run() {
  log_header ${FUNCNAME}

  if [ "$STAGE" != "" ] ; then
    reload_latest_versions
    if [ "$STAGE" == "release" ] ; then
      "$STAGE" "${DEVICE}"
    elif [ "$STAGE" == "rebuild_marlin_kernel" ] ; then
      if [ "${DEVICE}" == "marlin" ] || [ "${DEVICE}" == "sailfish" ]; then
        "$STAGE"
      fi
    else
      "$STAGE"
    fi
  else
    get_latest_versions
    check_for_new_versions
    aws_notify "RattlesnakeOS Build STARTED"
    setup_env
    check_chromium
    aosp_repo_init
    aosp_repo_modifications
    aosp_repo_sync
    aws_import_keys
    setup_vendor
    apply_patches
    # only marlin and sailfish need kernel rebuilt so that verity_key is included
    if [ "${DEVICE}" == "marlin" ] || [ "${DEVICE}" == "sailfish" ]; then
      rebuild_marlin_kernel
    fi
    build_aosp
    release "${DEVICE}"
    aws_upload
    checkpoint_versions
    aws_notify "RattlesnakeOS Build SUCCESS"
  fi
}
fi

full_run
`

	return txt, nil
}

var output = flag.String("output", "stack-builder", "Output file for stack script.")
var device = flag.String("device", "marlin", "build the stack for this device")
var releaseDownloadAddress = flag.String("release-download-address", "", "URL where the Android platform will look for published updates")
var buildType = flag.String("build-type", "user", "build type (user or userdebug)")
var chromiumVersion = flag.String("chromium-version", "", "build with a specific version of Chromium")
var hostsFileUrl = flag.String("hosts-file-url", "", "build with a custom hosts file from an URL")
var ignoreVersionChecks = flag.Bool("ignore-version-checks", false, "ignore version checks altogether, building again")
var customConfig = flag.String("custom-config", "", "path to a JSON file that has customizations (patches, script, prebuilts, et cetera) in the same AWSStackConfig structure documented in https://github.com/dan-v/rattlesnakeos-stack/README.md -- only the Custom structure members are respected")

type myStackConfig struct {
	*stack.AWSStackConfig
	BuildType              string
	ReleaseDownloadAddress string
}

func main() {
	flag.Parse()
	customizations := stack.AWSStackConfig{}
	if *customConfig != "" {
		contents, err := ioutil.ReadFile(*customConfig)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(contents, &customizations)
		if err != nil {
			panic(err)
		}
	}
	ignored := "ignored"
	preconfig := &stack.AWSStackConfig{
		Name:                   "rattlesnakeos",
		Region:                 ignored,
		AMI:                    ignored,
		Email:                  ignored,
		InstanceType:           ignored,
		InstanceRegions:        ignored,
		SkipPrice:              ignored,
		MaxPrice:               ignored,
		Version:                ignored,
		SSHKey:                 ignored,
		Schedule:               ignored,
		Device:                 *device,
		ChromiumVersion:        *chromiumVersion,
		IgnoreVersionChecks:    *ignoreVersionChecks,
		HostsFile:              *hostsFileUrl,
		EncryptedKeys:          false,
		CustomPatches:          customizations.CustomPatches,
		CustomScripts:          customizations.CustomScripts,
		CustomPrebuilts:        customizations.CustomPrebuilts,
		CustomManifestRemotes:  customizations.CustomManifestRemotes,
		CustomManifestProjects: customizations.CustomManifestProjects,
	}
	config := &myStackConfig{
		AWSStackConfig:         preconfig,
		BuildType:              *buildType,
		ReleaseDownloadAddress: *releaseDownloadAddress,
	}

	modded, err := alterTemplate(templates.BuildTemplate)
if err != nil {
panic(err)
}

	renderedBuildScript, err := stack.RenderTemplate(modded, config)
	if err != nil {
		log.Fatalf("Failed to render build script: %v", err)
	}
	log.Printf("Script that will run:\n==================================================%s\n==================================================", string(renderedBuildScript))

	cmd := []string{*output, *device}
	configStr, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		panic(err)
	}

	log.Printf("Settings that will be used:\n%s", string(configStr))
	log.Printf("Command prefix that will run: %s", cmd)

	err = ioutil.WriteFile(*output, renderedBuildScript, 0755)
	if err != nil {
		panic(err)
	}
}
