package appenginetesting

import (
    "text/template"
)

const appYAMLTemplString = `
application: {{.AppId}} 
version: 1
runtime: go
api_version: {{.APIVersion}}

handlers:
- url: /.*
  script: _go_app
`

var appYAMLTempl = template.Must(template.New("app.yaml").Parse(appYAMLTemplString))
