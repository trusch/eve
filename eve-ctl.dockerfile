FROM alpine
COPY eve-ctl.amd64 /bin/eve-ctl
CMD /bin/eve-ctl
