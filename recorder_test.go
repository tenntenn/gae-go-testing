package appenginetesting

import (
    "net/http"
    "net/http/httptest"
    "fmt"
    "testing"
    "bytes"
    "appengine"
    "appengine/datastore"
)

// Using wrapper for unit test.
var contextCreator func(r *http.Request) appengine.Context = appengine.NewContext

func sampleHandler(w http.ResponseWriter, r *http.Request) {
    c := contextCreator(r)
    k := datastore.NewKey(c, "Entity", "", 1, nil)
    e := &Entity{Foo:"foo", Bar:"bar"}
    _, err := datastore.Put(c, k, e)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprint(w, "OK")
}

func TestHandler(t *testing.T) {
    r, _ := http.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()

    recorder := NewContextRecorder(nil)
    contextCreator = recorder.Creator()
    sampleHandler(w, r)
    defer recorder.Context().Close()

    body := w.Body.Bytes()
    if !bytes.Equal(body, []byte("OK")) {
        t.Errorf("got response %v ; want %v", body, []byte("OK"))
    }

    var e Entity
    c := recorder.Context()
    k := datastore.NewKey(c, "Entity", "", 1, nil)
    if err := datastore.Get(c, k, &e); err != nil {
        t.Errorf(err.Error())
    }
    if e.Foo != "foo" || e.Bar != "bar" {
        t.Errorf("got response %v ; want %v", e, Entity{Foo:"foo", Bar:"bar"})
    }
}
