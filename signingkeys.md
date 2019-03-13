# Signing keys generation

Now note the device you'll build images for (e.g., `marlin`).  We'll use this shortly.

Create the keys now, using a secure machine that won't leak your keys.

## Android Verified Boot 1.0 (`marlin` generation devices)

```
mkdir -p keys/marlin
cd keys/marlin
../../development/tools/make_key releasekey '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key platform '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key shared '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key media '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key verity '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
cd ../..
```

## Android Verified Boot 2.0 (`taimen` generation devices)

```
mkdir -p keys/taimen
cd keys/taimen
../../development/tools/make_key releasekey '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key platform '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key shared '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
../../development/tools/make_key media '/C=US/ST=California/L=San Francisco/O=Your Business/OU=Your Business/CN=Your Business/emailAddress=webmaster@yourbusiness.com'
openssl genrsa -out avb.pem 2048
../../external/avb/avbtool extract_public_key --key avb.pem --output avb_pkmd.bin
cd ../..
```

Be sure to back up and safeguad these keys.  If you lose the keys, you won't be able to create new flashable builds without unlocking and wiping your device.  If your keys are stolen, someone could upload a malicious ROM to your device without you noticing.
