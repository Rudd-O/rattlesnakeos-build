package main

import (
	"./templates"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

var output = flag.String("output", "stack-builder", "Output file for stack script.")
var forceBuild = flag.Bool("force-build", false, "Force build even if no new versions exist of components.")
var skipChromiumBuild = flag.Bool("skip-chromium-build", false, "Skip Chromium build if Chromium was already built.")
var releaseUrl = flag.String("release-url", "http://example.com/", "Release URL.")
var buildType = flag.String("build-type", "user", "Which build type to use.")
var repoPatches = flag.String("repo-patches", "", "An advanced option that allows you to specify a git repo with patches to apply to AOSP build tree. see https://github.com/RattlesnakeOS/community_patches for more details.")
var repoPrebuilts = flag.String("repo-prebuilts", "", "An advanced option that allows you to specify a git repo with prebuilt APKs. see https://github.com/RattlesnakeOS/example_prebuilts for more details.")
var hostsFile = flag.String("hosts-file", "", `An advanced option that allows you to specify a replacement /etc/hosts file to enable global dns adblocking (e.g. https://raw.
githubusercontent.com/StevenBlack/hosts/master/hosts). note: be careful with this, as you 1) won't get any sort of notification on blocking 2) if you need to unblock something you'll have to rebuild the OS`)

type Data struct {
	Region            string
	Version           string
	PreventShutdown   string
	Force             string
	SkipChromiumBuild string
	Name              string
	RepoPatches       string
	RepoPrebuilts     string
	HostsFile         string
}

func boolStr(b bool) string {
	bStr := "false"
	if b {
		bStr = "true"
	}
	return bStr
}

func replace(text string, original string, substitution string, numReplacements int) (string, error) {
	newText := strings.Replace(text, original, substitution, numReplacements)
	if text == newText {
		return "", fmt.Errorf("The replacement of %s for %s produced no changes", original, substitution)
	}
	return newText, nil
}

func main() {
	flag.Parse()
	txt := templates.BuildTemplate
	var err error

	var replacements = []struct {
		original        string
		substitution    string
		numReplacements int
	}{
		{"<%", "{{", -1},
		{"%>", "}}", -1},
		{
			`AWS_SNS_ARN=$(aws --region ${REGION} sns list-topics --query 'Topics[0].TopicArn' --output text | cut -d":" -f1,2,3,4,5)":${STACK_NAME}"`,
			`AWS_SNS_ARN=none`,
			-1,
		},
		{`$(curl -s http://169.254.169.254/latest/meta-data/instance-type)`, "none", -1},
		{`$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | awk -F\" '/region/ {print $4}')`, "none", -1},
		{`$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)`, "none", -1},
		{`echo "No build is required, but FORCE_BUILD=true"`, `aws_notify "No build is required, but FORCE_BUILD=true"`, -1},
		{`echo "New build is required"`, `aws_notify "New build is required"`, -1},
		{
			`BUILD_TYPE="user"`,
			fmt.Sprintf(`BUILD_TYPE="%s" # replaced`, *buildType),
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
  quiet gitcleansources`,
			-1,
		},
		{
			`mkdir -p ${BUILD_DIR}/external/chromium/prebuilt/arm64`,
			``,
			-1,
		},
		{
			`cp out/Default/apks/MonochromePublic.apk ${BUILD_DIR}/external/chromium/prebuilt/arm64/`,
			`aws s3 cp out/Default/apks/MonochromePublic.apk "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk"`,
			-1,
		},
		{
			`aws s3 cp "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk" ${BUILD_DIR}/external/chromium/prebuilt/arm64/`,
			`# Suppressed copy from S3 to external/prebuilt/arm64/ as this happens later`,
			-1,
		},
		{`fetch --nohooks android`, `test -f .gclient || fetch --nohooks android`, -1},
		{
			"yes | gclient sync --with_branch_heads --jobs 32 -RDf",
			`quiet gitcleansources
  yes | gclient sync --with_branch_heads --jobs 32 -RDf`,
			-1,
		},
		{`out/Default`, `"$HOME"/chromium-out`, -1},
		{`rm -rf $HOME/chromium`, ``, -1},
		{
			"linux-image-$(uname --kernel-release)",
			"$(apt-cache search linux-image-* | awk ' { print $1 } ' | sort | egrep -v -- '(-dbg|-rt|-pae)' | grep ^linux-image-[0-9][.] | tail -1)",
			-1,
		},
		{
			`retry git clone "${KERNEL_SOURCE_URL}" "${MARLIN_KERNEL_SOURCE_DIR}"`,
			`gitavoidreclone "${KERNEL_SOURCE_URL}" "${MARLIN_KERNEL_SOURCE_DIR}"`,
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
    ln --verbose --symbolic ${BUILD_DIR}/keys/${DEVICE}/verity_user.der.x509 ${MARLIN_KERNEL_SOURCE_DIR}/verity_user.der.x509;
    cd ${MARLIN_KERNEL_SOURCE_DIR};
    make -j$(nproc --all) ARCH=arm64 marlin_defconfig;
    make -j$(nproc --all) ARCH=arm64 CONFIG_COMPAT_VDSO=n CROSS_COMPILE=${BUILD_DIR}/prebuilts/gcc/linux-x86/aarch64/aarch64-linux-android-4.9/bin/aarch64-linux-android-;
    cp -f arch/arm64/boot/Image.lz4-dtb ${BUILD_DIR}/device/google/marlin-kernel/;
    rm -rf ${BUILD_DIR}/out/build_*;
  "`,
			`bash -c "\
    set -ex ;
    mkdir -p ${MARLIN_KERNEL_OUT_DIR} ;
    cd ${BUILD_DIR};
    set +x
    . build/envsetup.sh;
    set -x
    make -j$(nproc --all) dtc mkdtimg;
    export PATH=${BUILD_DIR}/out/host/linux-x86/bin:${PATH};
    ln --verbose --symbolic -f ${BUILD_DIR}/keys/${DEVICE}/verity_user.der.x509 ${MARLIN_KERNEL_SOURCE_DIR}/verity_user.der.x509;
    cd ${MARLIN_KERNEL_SOURCE_DIR} ;
    make -j$(nproc --all) ARCH=arm64 marlin_defconfig O=${MARLIN_KERNEL_OUT_DIR} || make -j$(nproc --all) ARCH=arm64 mrproper marlin_defconfig O=${MARLIN_KERNEL_OUT_DIR} ;
    make -j$(nproc --all) ARCH=arm64 CONFIG_COMPAT_VDSO=n CROSS_COMPILE=${BUILD_DIR}/prebuilts/gcc/linux-x86/aarch64/aarch64-linux-android-4.9/bin/aarch64-linux-android- O=${MARLIN_KERNEL_OUT_DIR} ;
    # Now copy the recently-built kernel from its kernel-out place.
    rsync -a --inplace ${MARLIN_KERNEL_OUT_DIR}/arch/arm64/boot/Image.lz4-dtb ${BUILD_DIR}/device/google/marlin-kernel/Image.lz4-dtb
    rm -rf ${BUILD_DIR}/out/build_*;
  "`,
			-1,
		},
		{
			`# make modifications to default AOSP`,
			`# make modifications to default AOSP
  # Since we just git cleaned everything, we will have to re-copy
  # the MonochromePublic.apk file from its chromium-out place.
  mkdir -p ${BUILD_DIR}/external/chromium/prebuilt/arm64
  aws s3 cp "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk" ${BUILD_DIR}/external/chromium/prebuilt/arm64/
  `,
			-1,
		},
		{
			`git clone $REPO_PATCHES ${patches_dir}`,
			`gitavoidreclone "$REPO_PATCHES" ${patches_dir}`,
			-1,
		},
		{
			`timeout 30m "${BUILD_DIR}/vendor/android-prepare-vendor/execute-all.sh" --debugfs --yes --device "${DEVICE}" --buildID "${AOSP_BUILD}" --output "${BUILD_DIR}/vendor/android-prepare-vendor"`,
			`mkdir -p "${HOME}/vendor-in"
  local flag="${HOME}/vendor-in/.${DEVICE}-$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")"
  if test -f "${flag}" ; then
    true
  else
    "${BUILD_DIR}/vendor/android-prepare-vendor/execute-all.sh" --fuse-ext2 --yes --device "${DEVICE}" --buildID "${AOSP_BUILD}" --output "${HOME}/vendor-in"
    touch "${flag}"
  fi`,
			-1,
		},
		{
			`  mkdir --parents "${BUILD_DIR}/vendor/google_devices" || true
  rm --recursive --force "${BUILD_DIR}/vendor/google_devices/$DEVICE" || true
  mv "${BUILD_DIR}/vendor/android-prepare-vendor/${DEVICE}/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/${DEVICE}" "${BUILD_DIR}/vendor/google_devices"

  # smaller devices need big brother vendor files
  if [ "$DEVICE" == 'sailfish' ]; then
    rm --recursive --force "${BUILD_DIR}/vendor/google_devices/marlin" || true
    mv "${BUILD_DIR}/vendor/android-prepare-vendor/sailfish/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/marlin" "${BUILD_DIR}/vendor/google_devices"
  fi
  if [ "$DEVICE" == 'walleye' ]; then
    rm --recursive --force "${BUILD_DIR}/vendor/google_devices/muskie" || true
    mv "${BUILD_DIR}/vendor/android-prepare-vendor/walleye/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/muskie" "${BUILD_DIR}/vendor/google_devices"
  fi
`,
			`  mkdir --parents "${BUILD_DIR}/vendor/google_devices"
  rsync -avHAX --inplace --delete --delete-excluded "${HOME}/vendor-in/${DEVICE}/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/${DEVICE}/" "${BUILD_DIR}/vendor/google_devices/${DEVICE}/"

  # smaller devices need big brother vendor files
  if [ "$DEVICE" == 'sailfish' ]; then
    rsync -avHAX --inplace --delete --delete-excluded "${HOME}/vendor-in/sailfish/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/marlin/" "${BUILD_DIR}/vendor/google_devices/marlin/"
  fi
  if [ "$DEVICE" == 'walleye' ]; then
    rsync -avHAX --inplace --delete --delete-excluded "${HOME}/vendor-in/walleye/$(tr '[:upper:]' '[:lower:]' <<< "${AOSP_BUILD}")/vendor/google_devices/marlin/" "${BUILD_DIR}/vendor/google_devices/muskie/"
  fi
`,
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
	}

	for _, r := range replacements {
		if txt, err = replace(txt, r.original, r.substitution, r.numReplacements); err != nil {
			log.Fatalf("%s", err)
		}
	}

	data := Data{
		Force:             boolStr(*forceBuild),
		SkipChromiumBuild: boolStr(*skipChromiumBuild),
		Name:              "rattlesnakeos",
		Region:            "none",
		RepoPatches:       *repoPatches,
		RepoPrebuilts:     *repoPrebuilts,
		HostsFile:         *hostsFile,
	}

	t, err := template.New("stack").Parse(txt)
	if err != nil {
		log.Fatalf("%s", err)
	}

	var tpl bytes.Buffer
	if err = t.Execute(&tpl, data); err != nil {
		log.Fatalf("%s", err)
	}

	s := tpl.String()
	if strings.Contains(s, "<%") {
		s = strings.Split(tpl.String(), "<%")[0]
		s = s + "<%" + strings.Split(tpl.String(), "<%")[1]
		log.Fatalf("The resultant string did not render properly.\n\n %s", s)
	}
	s = strings.Split(s, "\nfull_run\n")[0]
	s = s + `
aws() {
  func="$1"
  cmd="$2"
  in="$3"
  out="$4"
  if [ "$func" == "sns" ]
  then
	if [[ $7 == --message=* ]]
	then
		echo "${7#--message=}" | sed 's/^/aws_notify: /' >&2
	else
		echo "$8" | sed 's/^/aws_notify: /' >&2
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

gen_keys() {
	echo "This program needs the keys already present in s3://${AWS_KEYS_BUCKET}/${DEVICE}" >&2
	false
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

gitcleansources() {
	local gitstatus
	local r
	local gitdir
	for gitdir in $(find -name .git -type d) ; do
		pushd "$gitdir/.." > /dev/null || continue
		gitstatus=$(git status --ignored) || { r=$? ; echo "$gitstatus" >&2 ; popd > /dev/null ; return $? ; }
		if echo "$gitstatus" | grep -q "nothing to commit, working tree clean" && echo "$gitstatus" | grep -qv "Untracked files:" && echo "$gitstatus" | grep -qv "Ignored files:" ; then
			true
		else
			pwd
			echo "$gitstatus" >&2
			git clean -fxd || { r=$? ; popd > /dev/null ; return $r ; }
			git reset --hard || { r=$? ; popd > /dev/null ; return $r ; }
		fi
		popd > /dev/null
	done
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

aws_import_keys() {
  log_header ${FUNCNAME}
  if [ "$(aws s3 ls "s3://${AWS_KEYS_BUCKET}/${DEVICE}" | wc -l)" == '0' ]; then
    echo "Keys do not exist for ${DEVICE}" >&2
    return 1
  else
    log "Keys already exist for ${DEVICE} - grabbing them from S3"
    mkdir -p "${BUILD_DIR}/keys"
    aws s3 sync "s3://${AWS_KEYS_BUCKET}" "${BUILD_DIR}/keys"

    # Must copy this here because aws_gen_keys is short-circuited.
    if [ "${DEVICE}" == "marlin" ] || [ "${DEVICE}" == "sailfish" ]; then
      gen_verity_key "${DEVICE}"
    fi

    if [ "${DEVICE}" == "walleye" ] || [ "${DEVICE}" == "taimen" ]; then
      gen_avb_key "${DEVICE}"
    fi
  fi
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
    setup_vendor
    aws_import_keys
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
`
	s = s + "\nRELEASE_URL=" + *releaseUrl
	s = s + "\nfull_run\n"
	err = ioutil.WriteFile(*output, []byte(s), 0755)
	if err != nil {
		log.Fatalf("%s", err)
	}
}
