#!/bin/bash

acbuild --debug begin

# In the event of the script exiting, end the build
trap "{ export EXT=$?; acbuild --debug end && exit $EXT; }" EXIT

acbuild --debug set-name trusch.io/eve-ctl
acbuild --debug label add version $(git describe)
acbuild --debug dependency add trusch.io/alpine
acbuild --debug copy bin/eve-ctl.amd64 /bin/eve-ctl
acbuild --debug set-exec -- /bin/eve-ctl
acbuild --debug write --overwrite bin/eve-ctl-$(git describe)-amd64.aci

gpg --sign --armor --detach bin/eve-ctl-$(git describe)-amd64.aci

ln -sf eve-ctl-$(git describe)-amd64.aci bin/eve-ctl-latest-amd64.aci
ln -sf eve-ctl-$(git describe)-amd64.aci.asc bin/eve-ctl-latest-amd64.aci.asc

exit $?
