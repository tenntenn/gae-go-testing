// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// This file changed by Takuya Ueda from http://code.google.com/p/gae-go-testing/.

// Package appenginetesting provides an appengine.Context for testing.
package appenginetesting

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"code.google.com/p/goprotobuf/proto"

	"appengine"
	"appengine_internal"
)

// Statically verify that Context implements appengine.Context.
var _ appengine.Context = (*Context)(nil)

// httpClient is used to communicate with the helper child process's
// webserver.  We can't use http.DefaultClient anymore, as it's now
// blacklisted in App Engine 1.6.1 due to people misusing it in blog
// posts and such.  (but this is one of the rare valid uses of not
// using urlfetch)
var httpClient = &http.Client{}


// Default API Version
const DefaultAPIVersion = "go1"

// API version of golang.
// It is used for app.yaml of dev_server setting.
var APIVersion = DefaultAPIVersion

// Context implements appengine.Context by running a dev_appserver.py
// process as a child and proxying all Context calls to the child.
// Use NewContext to create one.
type Context struct {
	appid  string
	req    *http.Request
	child  *exec.Cmd
	port   int    // of child dev_appserver.py http server
	appDir string // temp dir for application files
}

func (c *Context) AppID() string {
	return c.appid
}

func (c *Context) logf(level, format string, args ...interface{}) {
	log.Printf(level+": "+format, args...)
}

func (c *Context) Debugf(format string, args ...interface{})    { c.logf("DEBUG", format, args...) }
func (c *Context) Infof(format string, args ...interface{})     { c.logf("INFO", format, args...) }
func (c *Context) Warningf(format string, args ...interface{})  { c.logf("WARNING", format, args...) }
func (c *Context) Errorf(format string, args ...interface{})    { c.logf("ERROR", format, args...) }
func (c *Context) Criticalf(format string, args ...interface{}) { c.logf("CRITICAL", format, args...) }

func (c *Context) Call(service, method string, in, out proto.Message, opts *appengine_internal.CallOptions) error {
	data, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("http://127.0.0.1:%d/call?s=%s&m=%s", c.port, service, method),
		bytes.NewBuffer(data))
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("got status %d; body: %q", res.StatusCode, body)
	}
	pbytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return proto.Unmarshal(pbytes, out)
}

func (c *Context) FullyQualifiedAppID() string {
	// TODO(bradfitz): is this right, prepending "dev~"?  It at
	// least appears to make the Python datastore fake happy.
	return "dev~" + c.appid
}

func (c *Context) Request() interface{} {
	return c.req
}

// Close kills the child dev_appserver.py process, releasing its
// resources.
//
// Close is not part of the appengine.Context interface.
func (c *Context) Close() {
	if c.child == nil {
		return
	}
	if p := c.child.Process; p != nil {
		p.Signal(syscall.SIGTERM)
	}
	os.RemoveAll(c.appDir)
	c.child = nil
}

// Options control optional behavior for NewContext.
type Options struct {
	// AppId to pretend to be. By default, "testapp"
	AppId string
}

func (o *Options) appId() string {
	if o == nil || o.AppId == "" {
		return "testapp"
	}
	return o.AppId
}

func findFreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findDevAppserver() (string, error) {
	if e := os.Getenv("APPENGINE_SDK"); e != "" {
		p := filepath.Join(e, "dev_appserver.py")
		if fileExists(p) {
			return p, nil
		}
		return "", fmt.Errorf("invalid APPENGINE_SDK environment variable; path %q doesn't exist", p)
	}
	try := []string{
		filepath.Join(os.Getenv("HOME"), "sdk", "go_appengine", "dev_appserver.py"),
		filepath.Join(os.Getenv("HOME"), "sdk", "google_appengine", "dev_appserver.py"),
		filepath.Join(os.Getenv("HOME"), "google_appengine", "dev_appserver.py"),
		filepath.Join(os.Getenv("HOME"), "go_appengine", "dev_appserver.py"),
	}
	for _, p := range try {
		if fileExists(p) {
			return p, nil
		}
	}
	return exec.LookPath("dev_appserver.py")
}

func (c *Context) startChild() error {

	port, err := findFreePort()
	if err != nil {
		return err
	}

	c.appDir, err = ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	err = os.Mkdir(filepath.Join(c.appDir, "helper"), 0755)
	if err != nil {
		return err
	}

    appYAMLBuf := new(bytes.Buffer)
    appYAMLTempl.Execute(appYAMLBuf, struct {
        AppId string
        APIVersion string
    }{
        c.appid,
        APIVersion,
    })
	err = ioutil.WriteFile(filepath.Join(c.appDir, "app.yaml"), appYAMLBuf.Bytes(), 0755)
	if err != nil {
		return err
	}

    helperBuf := new(bytes.Buffer)
    helperTempl.Execute(helperBuf, nil)
	err = ioutil.WriteFile(filepath.Join(c.appDir, "helper", "helper.go"), helperBuf.Bytes(), 0644)
	if err != nil {
		return err
	}

	devAppserver, err := findDevAppserver()

	c.port = port
	c.child = exec.Command(
		devAppserver,
		"--clear_datastore",
		"--high_replication",
		// --blobstore_path=... <tempdir>
		// --datastore_path=DS_FILE
		"--skip_sdk_update_check",
		fmt.Sprintf("--port=%d", port),
		c.appDir,
	)
	stderr, err := c.child.StderrPipe()
	if err != nil {
		return err
	}

	err = c.child.Start()
	if err != nil {
		return err
	}

	r := bufio.NewReader(stderr)
	donec := make(chan bool)
	errc := make(chan error)
	go func() {
		done := false
		for {
			bs, err := r.ReadSlice('\n')
			if err != nil {
				errc <- err
				return
			}
			line := string(bs)
			if done {
				// Uncomment for extra debugging, to see what the child is logging.
				//log.Printf("child: %q", line)
				continue
			}
			if strings.Contains(line, "Running application") {
				done = true
				donec <- true
			}
		}
	}()

	select {
	case err := <-errc:
		return fmt.Errorf("error starting child process: %v", err)
	case <-time.After(10e9):
		if p := c.child.Process; p != nil {
			p.Kill()
		}
		return errors.New("timeout starting process")
	case <-donec:
	}

    return nil
}

// NewContext returns a new AppEngine context with an empty datastore, etc.
// A nil Options is valid and means to use the default values.
func NewContext(opts *Options) (*Context, error) {
	req, _ := http.NewRequest("GET", "/", nil)
	c := &Context{
		appid: opts.appId(),
		req:   req,
	}
	if err := c.startChild(); err != nil {
		return nil, err
	}
	return c, nil
}
