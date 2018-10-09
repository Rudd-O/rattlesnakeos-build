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
			`wget ${ANDROID_SDK_URL} -O sdk-tools.zip
  unzip sdk-tools.zip`,
			`if [ ! -f sdk-tools.zip ] ; then
	wget ${ANDROID_SDK_URL} -O sdk-tools.zip
fi
  unzip -o sdk-tools.zip  || {
	echo unzip failed, retrying download
	rm -f sdk-tools.zip
	wget ${ANDROID_SDK_URL} -O sdk-tools.zip
	unzip -o sdk-tools.zip
  }`,
			-1,
		},
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
  gitcleansources`,
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
			` aws s3 cp "${BUILD_DIR}/external/chromium/prebuilt/arm64/MonochromePublic.apk" "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk"`,
			``,
			-1,
		},
		{
			`# make modifications to default AOSP`,
			`# make modifications to default AOSP
  # Since we just git cleaned everything, we will have to re-copy
  # the MonochromePublic.apk file once again.
  mkdir -p ${BUILD_DIR}/external/chromium/prebuilt/arm64
  aws s3 cp "s3://${AWS_RELEASE_BUCKET}/chromium/MonochromePublic.apk" ${BUILD_DIR}/external/chromium/prebuilt/arm64/`,
			-1,
		},
		{`fetch --nohooks android`, `test -f .gclient || fetch --nohooks android`, -1},
		{
			"yes | gclient sync --with_branch_heads --jobs 32 -RDf",
			`gitcleansources  yes | gclient sync --with_branch_heads --jobs 32 -RDf`,
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
			`git clone "${KERNEL_SOURCE_URL}" "${MARLIN_KERNEL_SOURCE_DIR}"`,
			`if test -d "${MARLIN_KERNEL_SOURCE_DIR}"/.git ; then
	pushd "${MARLIN_KERNEL_SOURCE_DIR}"
	sed -i 's|url = .*|url = '"${KERNEL_SOURCE_URL}"'|' .git/config
	git fetch
	popd
  else
	git clone "${KERNEL_SOURCE_URL}" "${MARLIN_KERNEL_SOURCE_DIR}"
  fi`,
			-1,
		},
		{
			`bash -c "`,
			`bash -exc "`,
			-1,
		},
		{
			`ln --verbose --symbolic`,
			`ln --verbose --symbolic -f`,
			-1,
		},
		{
			`"${BUILD_DIR}/script/release.sh" "$DEVICE"`,
			`bash -x "${BUILD_DIR}/script/release.sh" "$DEVICE"`,
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
			cp -f "$in" "$out"
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
			git clean -dff || { r=$? ; popd > /dev/null ; return $r ; }
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

if [ "$ONLY_REPORT" == "true" ]
then
full_run() {
  get_latest_versions
  persist_latest_versions
  check_for_new_versions
}
else
full_run() {
  if [ "$STAGE" != "" ] ; then
    reload_latest_versions
    if [ "$STAGE" == "rebuild_marlin_kernel" ] ; then
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
    fetch_aosp_source
    setup_vendor
    aws_import_keys
    apply_patches
    # only marlin and sailfish need kernel rebuilt so that verity_key is included
    if [ "${DEVICE}" == "marlin" ] || [ "${DEVICE}" == "sailfish" ]; then
      rebuild_marlin_kernel
    fi
    build_aosp
    aws_release
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
