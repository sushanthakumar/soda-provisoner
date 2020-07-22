# Based on ubuntu
FROM ubuntu:xenial
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI block Plugin"

# Copy opensdsplugin from build directory
COPY soda-provisioner /soda-provisioner
# Define default command
ENTRYPOINT ["/soda-provisioner"]
