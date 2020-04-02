package gateway

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/frame/util"
)

const (
	head = `
	<link rel="stylesheet" href="https://unpkg.com/purecss@1.0.1/build/pure-min.css" integrity="sha384-oAOxQR6DkCoMliIh8yFnu25d7Eq/PHS21PClpwjOTeU2jRSq11vu66rf90/cZr47" crossorigin="anonymous">
	<style type="text/css">
	body {padding: 5px}
	</style>
	`
)

type GroupApi struct {
	Name string
	Apis []*API
}

func ApiList(c *gin.Context) {

	html := `
	<html>
		<head>
			%s
		</head>
		<body>
			<h3>API List</h3>

			<table class="pure-table pure-table-bordered">
			<thead><tr><th>Group</th><th>API</th><th>Description</th></tr><thead>
			<tbody>%s</tbody>
			</table>
		</body>
	</html>
	`

	apiGroupTemplate := `<tr><td rowspan="%d">%s</td><td><a href="./detail?name=%s">%s</a></td><td>%s</td>`
	apiTemplate := `<tr><td><a href="./detail?name=%s">%s</a></td><td>%s</td>`

	groups := make([]GroupApi, 0)
	for _, api := range apis {
		var hitGroup *GroupApi
		for i, group := range groups {
			if api.Group == group.Name {
				hitGroup = &groups[i]
			}
		}

		if hitGroup != nil {
			hitGroup.Apis = append(hitGroup.Apis, api)
			continue
		}

		group := GroupApi{
			Name: api.Group,
			Apis: []*API{api},
		}
		groups = append(groups, group)
	}

	sb := strings.Builder{}

	for _, g := range groups {
		for i, api := range g.Apis {
			if i == 0 {
				sb.WriteString(fmt.Sprintf(apiGroupTemplate, len(g.Apis), api.Group, api.Group+"."+api.Key, url.QueryEscape(api.Key), api.Name))

			} else {
				sb.WriteString(fmt.Sprintf(apiTemplate, api.Group+"."+api.Key, url.QueryEscape(api.Key), api.Name))
			}

		}
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(html, head, sb.String())))

}

func ApiDetail(c *gin.Context) {

	name := c.Request.URL.Query().Get("name")
	if len(name) == 0 {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte("<h1>No api name param</h1>"))
		return
	}

	var api *API
	for _, item := range apis {
		if item.Group+"."+item.Key == name {
			api = item
			break
		}
	}

	if api == nil {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte("<h1>Api not found</h1>"))
		return
	}

	html := `
	<html>
	<head>
		%s
	</head>
	<body>
		<h1>%s<h1><hr/>
		<h3>Mock:</h3>
		<code class="pre">%s</code>
		<div><span> %s </span>
		<h3>Request</h3>
		%s
		<h3>Response</h3>
		%s
	</body>
	</html>
	`

	request := strings.Builder{}
	request.WriteString(`
	<table class="pure-table pure-table-bordered">
	<thead><tr><th>FieldName</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
	<tbody>
	`)
	fieldTemplate := `<tr><td>%s</td><td>%s</td><td>%t</td><td>%s</td></tr>`
	for _, f := range api.Request.fields {
		request.WriteString(fmt.Sprintf(fieldTemplate, f.name, f.typ, f.required, f.desc))
	}
	request.WriteString("</tbody></table>")

	response := strings.Builder{}
	response.WriteString(`
	<table class="pure-table pure-table-bordered">
	<thead><tr><th>FieldName</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
	<tbody>
	`)
	for _, f := range api.Response.fields {
		response.WriteString(fmt.Sprintf(fieldTemplate, f.name, f.typ, f.required, f.desc))
	}
	response.WriteString("</tbody></table>")

	for _, typ := range api.Response.types {
		response.WriteString(fmt.Sprintf("<h4>%s</h4>", typ.name))
		response.WriteString(`
		<table class="pure-table pure-table-bordered">
		<thead><tr><th>FieldName</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
		<tbody>
		`)
		for _, f := range typ.fields {
			response.WriteString(fmt.Sprintf(fieldTemplate, f.name, f.typ, f.required, f.desc))
		}
		response.WriteString("</tbody></table>")
	}

	schema := "http"
	if c.Request.URL.Scheme == "https" {
		schema = "https"
	}

	path := fmt.Sprintf("%s://%s/api/%s/%s", schema, c.Request.Host, api.Group, strings.ReplaceAll(api.Key, ".", "/"))

	req := reflect.New(api.ReqType).Interface()
	err := faker.FakeData(&req)
	util.CheckError(err)
	body, err := Json.Marshal(&req)
	util.CheckError(err)

	mock := fmt.Sprintf(`<pre class="code">curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" -H "Mock: true" \
	'%s' \
	-d'%s'</pre>`, path, string(body))

	data := fmt.Sprintf(html, head, api.Name, mock, api.Summary, request.String(), response.String())

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(data))

}