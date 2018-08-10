package main

import (
	"./templates"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
)

var output = flag.String("output", "stack-builder", "Output file for stack script.")
var forceBuild = flag.Bool("force-build", false, "Force build even if no new versions exist of components.")
var patchChromium = flag.Bool("patch-chromium", true, "Patch Chromium with Bromite.")
var releaseUrl = flag.String("release-url", "http://example.com/", "Release URL.")
var buildType = flag.String("build-type", "user", "Which build type to use.")

type Data struct {
	Region          string
	Version         string
	PreventShutdown string
	Force           string
	PatchChromium   string
	Name            string
}

func main() {
	flag.Parse()
	txt := templates.ShellScriptTemplate
	txt = strings.Replace(txt, "<%", "{{", -1)
	txt = strings.Replace(txt, "%>", "}}", -1)
	txt = strings.Replace(txt,
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
		-1)
	txt = strings.Replace(txt,
		"git clone --branch ${CHROMIUM_REVISION} $BROMITE_URL $HOME/bromite",
		`if [ -d $HOME/bromite ] ; then
    pushd $HOME/bromite
    git fetch origin
    git checkout -f ${CHROMIUM_REVISION}
    popd
else
    git clone --branch ${CHROMIUM_REVISION} $BROMITE_URL $HOME/bromite
fi`,
		-1)
	if *buildType != "user" {
		txt = strings.Replace(txt, `"release aosp_${DEVICE} user"`, fmt.Sprintf(`"release aosp_${DEVICE} %s"`, *buildType), -1)
	}
	txt = strings.Replace(txt, "Stack Version: %s %s\\n  ", "", -1)
	txt = strings.Replace(txt, `"${STACK_VERSION}" "${STACK_UPDATE_MESSAGE}" `, "", -1)
	txt = strings.Replace(
		txt,
		"yes | gclient sync --with_branch_heads --jobs 32 -RDf",
		`for gitdir in $( find -name .git ) ; do
	pushd $gitdir/..
	git clean -dff
	popd
  done
  yes | gclient sync --with_branch_heads --jobs 32 -RDf`,
		-1)
	txt = strings.Replace(txt, "linux-image-$(uname --kernel-release)", "$(apt-cache search linux-image-* | awk ' { print $1 } ' | sort | egrep -v -- '(-dbg|-rt|-pae)' | grep ^linux-image-[0-9][.] | tail -1)", -1)
	txt = strings.Replace(
		txt,
		`git clone "${KERNEL_SOURCE_URL}" "${MARLIN_KERNEL_SOURCE_DIR}"`,
		`if test -d "${MARLIN_KERNEL_SOURCE_DIR}"/.git ; then
	pushd "${MARLIN_KERNEL_SOURCE_DIR}"
	sed -i 's|url = .*|url = '"${MARLIN_KERNEL_SOURCE_DIR}"'|' .git/config
	git fetch
	popd
  else
	git clone "${KERNEL_SOURCE_URL}" "${MARLIN_KERNEL_SOURCE_DIR}"
  fi`,
		-1)
	txt = strings.Replace(
		txt,
		`"${BUILD_DIR}/script/release.sh" "$DEVICE"`,
		`bash -x "${BUILD_DIR}/script/release.sh" "$DEVICE"`,
		-1,
	)
	txt = strings.Replace(
		txt,
		`"$(wget -O - "${RELEASE_URL}/${DEVICE}-stable")"`,
		`"$(aws s3 cp "s3://${AWS_RELEASE_BUCKET}/${RELEASE_CHANNEL}" -)"`,
		-1,
	)
	t, err := template.New("stack").Parse(txt)
	if err != nil {
		panic(err)
	}

	forceBuildStr := "false"
	if *forceBuild {
		forceBuildStr = "true"
	}
	patchChromiumStr := "false"
	if *patchChromium {
		patchChromiumStr = "true"
	}
	data := Data{
		Force:         forceBuildStr,
		PatchChromium: patchChromiumStr,
		Name:          "rattlesnakeos",
		Region:        "none",
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		panic(err)
	}
	s := tpl.String()
	if strings.Contains(s, "<%") {
		s = strings.Split(tpl.String(), "<%")[0]
		s = s + "<%" + strings.Split(tpl.String(), "<%")[1]
		panic(fmt.Sprintf("The resultant string did not render properly.\n\n %s", s))
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
		panic(err)
	}
}
