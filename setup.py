#!/usr/bin/env python
# _*_ coding: utf-8 _*_

# This script copy packages which is necessary for this library to go distribusion.
# Please run this script before installation such as:
#   curl https://raw.github.com/tenntenn/gae-go-testing/master/setup.py | python 

import os
import sys
import re

def main():

    # Check $APPENGINE_SDK
    if os.environ.get("APPENGINE_SDK") is None:
        print >>sys.stderr,"""Error: Please set path to \
                              GAE distribustion as APPENGINE_SDK such as \
                              'export APPENGINE_SDK=/usr/local/google_appengine/'."""
        return

    # Check $GOROOT
    if os.environ.get("GOROOT") is None:
        print >>sys.stderr, """Error: Please set path to \
                               Go distribustion as GOROOT such as \
                               'export GOROOT=/usr/local/go/'."""
        return

    # Check $PATH
    if os.environ.get("PATH") is None \
        or re.search(os.environ.get("APPENGINE_SDK"),
                 os.environ.get("PATH")) is None:

        print >>sys.stderr, """Error: Please add $APPENGINE_SDK to \
                               $PATH such as 'export PATH=$PATH:$APPENGINE_SDK'."""
        return

    # Check previous version
    packages = ["appengine", "appengine_internal", "code.google.com/p/goprotobuf"]
    for pkg in packages:
        dst = "{0}/src/pkg/{1}".format(os.environ.get("GOROOT"), pkg)
        if os.path.exists(dst):
            print >>sys.stderr, "Error: {0} is already existing".format(dst)
            return 

    # Copy appengine to go distribustion
    for pkg in packages:
        src = "{0}/goroot/src/pkg/{1}/*".format(os.environ.get("APPENGINE_SDK"), pkg)
        dst = "{0}/src/pkg/{1}".format(os.environ.get("GOROOT"), pkg)
        os.mkdir(dst)
        cmd = "cp -r {0} {1}".format(src, dst)
        print cmd
        os.system(cmd)

if __name__ == "__main__":
    main()
