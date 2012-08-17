package appenginetesting

import (
    "appengine"
    "net/http"
)

type ContextRecorder struct {
    c *Context
    creator func(r *http.Request) appengine.Context
}

func NewContextRecorder(opts *Options) *ContextRecorder {

    recorder := new(ContextRecorder)

    creator := func(r *http.Request) appengine.Context {
        recorder.c = &Context{
		    appid: opts.appId(),
		    req:   r,
	    }

        if err := recorder.c.startChild(); err != nil {
            panic(err.Error())
	    }

        return recorder.c
    }

    recorder.creator = creator

    return recorder
}

func (r *ContextRecorder) Creator() func(r *http.Request) appengine.Context {
    return r.creator
}

func (r *ContextRecorder) Context() *Context {
    return r.c
}
