gae-go-testing
==============

Testing library for Go App Engine, giving you an appengine.Context fake that forwards to a dev_appserver.py child process.
This library is fixed for go1 based on http://code.google.com/p/gae-go-testing/ .

Installation
-----

Before installing this library, you have to install [appengine SDK](https://developers.google.com/appengine/downloads#Google_App_Engine_SDK_for_Go).
And copy appengine, appengine_internal and goprotobuf as followings :

    $cp -r $GAEROOT/goroot/src/pkg/appengine $GOROOT/src/pkg/
    $cp -r $GAEROOT/goroot/src/pkg/appengine_internal $GOROOT/src/pkg/
    $cp -r $GAEROOT/goroot/src/pkg/code.google.com/p/goprotobuf $GOROOT/src/pkg/code.google.com/p/

In addition set environment variables :

    export PATH=$PATH:$GAEROOT 
    export APPENGINE_SDK=$GAEROT

This library can be installed as following :

    $go get github.com/tenntenn/gae-go-testing


Usage
-----

context_test.go and recorder_test.go show an example of usage.
