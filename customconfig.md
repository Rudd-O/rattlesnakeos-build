# RattlesnakeOS ROM customization options (`CUSTOM_CONFIG`)

RattlesnakeOS has [a variety of customization options you can use in order to customize your build](https://github.com/dan-v/rattlesnakeos-stack#configuration).  However, there is a caveat: unlike RattlesnakeOS, the configuration options aren't supplied in the same way to the build.

In RattlesnakeOS, you ordinarily create a TOML configuration file `.rattlesnakeos.toml` using the `rattlesnakeos-stack config` command, then you edit and save the configuration file.  This configuration file is then used by `rattlesnakeos-stack deploy`.

When [building using Jenkins](jenkins.md), you can do one of three things:

1. Place the JSON text as the (string) value of the `CUSTOM_CONFIG` parameter of your `parameters.groovy` file.  This is what is shown above.
2. *Check in* the customization options as a JSON file named `custom-config.json`, alongside the `Jenkinsfile` within.  Yes, this requires you to fork this project to your own repository.
3. Manually paste JSON text directly into the *Build with parameters* page.  Any text pasted there will override any existing `custom-config.json` file for that specific build.  But the next build won't remember this action, so it will be done without the custom config you pasted.

When [building interactively](interactive.md), you can write your `custom-config.json` file, and pass it to the builder with `-custom-config custom-config.json` on the command line.

JSON being not TOML, there are syntax differences between what may go into the configuration file.  The semantics of the custom configuration, however, are the same.

Let's use the RattlesnakeOS README example for our purpose here.  Suppose you'd put this into a `.rattlesnakeos.toml` config:

```
[[custom-patches]]
  repo = "https://github.com/RattlesnakeOS/community_patches"
  patches = [
      "00001-global-internet-permission-toggle.patch", "00002-global-sensors-permission-toggle.patch",
  ]

[[custom-scripts]]
  repo = "https://github.com/RattlesnakeOS/example_patch_shellscript"
  scripts = [ "00002-custom-boot-animation.sh" ]
```

Here's how the *exact same thing* would look like, in your `custom-config.json` file:

```
{
    "custom-patches": [{
        "repo": "https://github.com/RattlesnakeOS/community_patches",
        "patches": [
            "00001-global-internet-permission-toggle.patch",
            "00002-global-sensors-permission-toggle.patch"
        ]
    }],
    "custom-scripts": [{
        "repo": "https://github.com/RattlesnakeOS/example_patch_shellscript",
        "scripts": [
            "00002-custom-boot-animation.sh"
        ]
    }]
}
```

See?  Nothing wow or extraordinary.  It's merely a format change.

One small caveat: this program supports only a limited subset of options of the `.rattlesnakeos.toml` configuration file.  Here is a comprehensive list:

```
		CustomPatches:          customizations.CustomPatches,
		CustomScripts:          customizations.CustomScripts,
		CustomPrebuilts:        customizations.CustomPrebuilts,
		CustomManifestRemotes:  customizations.CustomManifestRemotes,
		CustomManifestProjects: customizations.CustomManifestProjects,
```

*For the programming-curious:* The data structures populated by both `.rattlesnakeos.toml` and `custom-config.json` are defined in file https://github.com/dan-v/rattlesnakeos-stack/blob/9.0/stack/aws.go . A full reference to the ROM customization options is available under the [the Customizations section of the RattlesnakeOS README](https://github.com/dan-v/rattlesnakeos-stack).

*Note:* If you opted for the JSON-in-`parameters.groovy` option, have the Jenkins project *Scan Multibranch Pipeline Now*.  This causes the build to pick up the new defaults.  Cancel any build that happens as a result of the rescan, and manually dispatch one more build.  If you opted for the fork-and-check-in-my-own-`config.json` option, all you have to do is commit and push your changes — your build server will start to build.
