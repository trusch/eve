#!/bin/bash

acbuild --debug begin

# In the event of the script exiting, end the build
trap "{ export EXT=$?; acbuild --debug end && exit $EXT; }" EXIT

acbuild --debug set-name trusch.io/eve
acbuild --debug label add version $(git describe)
acbuild --debug dependency add trusch.io/alpine
acbuild --debug copy bin/eve.amd64 /bin/eve
acbuild --debug set-exec -- /bin/eve
acbuild --debug port add http tcp 80
acbuild --debug port add https tcp 443
acbuild --debug write --overwrite bin/eve-$(git describe)-amd64.aci

gpg --sign --armor --detach bin/eve-$(git describe)-amd64.aci

ln -sf eve-$(git describe)-amd64.aci bin/eve-latest-amd64.aci
ln -sf eve-$(git describe)-amd64.aci.asc bin/eve-latest-amd64.aci.asc

exit $?
