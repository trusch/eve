#!/bin/bash

acbuild --debug begin

# In the event of the script exiting, end the build
trap "{ export EXT=$?; acbuild --debug end && exit $EXT; }" EXIT

acbuild --debug set-name trusch.io/eve-ctl
acbuild --debug label add version $(git describe)
acbuild --debug label add arch armv7l
acbuild --debug dependency add trusch.io/alpine
acbuild --debug copy bin/eve-ctl.arm /bin/eve-ctl
acbuild --debug set-exec -- /bin/eve-ctl
acbuild --debug write --overwrite bin/eve-ctl-$(git describe)-armv7l.aci

gpg --sign --armor --detach bin/eve-ctl-$(git describe)-armv7l.aci

ln -sf eve-ctl-$(git describe)-armv7l.aci bin/eve-ctl-latest-armv7l.aci
ln -sf eve-ctl-$(git describe)-armv7l.aci.asc bin/eve-ctl-latest-armv7l.aci.asc

exit $?
