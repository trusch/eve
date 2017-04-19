FROM alpine
COPY bin/eve-ctl.amd64 /bin/eve-ctl
CMD /bin/eve-ctl
