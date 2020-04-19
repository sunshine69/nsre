#!/bin/sh
# This is to run in the docker container as the jenkins user
# This takes a long time so better build the image having the user jenkins and
# correct ownership only for running in jenkins, and another image if u need to
# build it localy to match your current user

# This is to illustrate how to run extra build script as different users using
# the jenkins helper groovy script.
if `grep jenkins /etc/passwd >/dev/null 2>&1`; then
    chown -R jenkins:jenkins /go
fi
