// out=`pwd`/api_handlers.go in=`pwd`/api.go go generate handlers_gen/codegen.go

//go:generate echo Start generate handlres for $in
//go:generate go run .
//go:generate go fmt $out
//go:generate echo Done! Ganerated code in $out

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

type apiParamsField struct {
	Name string
	Type string
	Tag  string
}
type funcOption struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}
type apiFunc struct {
	Name        string
	BindToName  string
	Options     funcOption
	InTypeName  string
	OutTypeName string
}
type apiHandler struct {
	Path   string
	Method string
	Name   string
}

type TagData struct {
	Name            string
	Value           string
	ValuesList      string
	QuotaValuesList string
	DefaultValue    string
}
type FieldData struct {
	Type      string
	Name      string
	ParamName string
	Tags      []TagData
}
type FillData struct {
	Name   string
	Fields []FieldData
}
type HandleRequestData struct {
	BindToName      string
	HandlerName     string
	Auth            bool
	ParamsName      string
	FuncName        string
	ResponseType    string
	ErrorJSONTag    string
	ResponseJSONTag string
}

var (
	HeadTmpl = `
package main

import (
	{{ range .Imports }} "{{ . }}" 
	{{ end }}
)

type ErrorApi struct {
	Error string {{ .ErrorJSONTag }}
}`

	FillTmpl = `
func (p *{{ .Name }}) Fill(r *http.Request) error {
{{- range .Fields }}
	{{ template "FieldTmpl" . }}
	p.{{.Name}} = {{.ParamName}}
{{- end }}
	return nil
}`
	FieldTmpl = `
{{- if eq .Type "string" -}}
	{{.ParamName}} := r.FormValue("{{.ParamName}}")
{{- else -}}
{{.ParamName}}, err := strconv.Atoi(r.FormValue("{{.ParamName}}"))
	if err != nil {
		return errors.New("{{.ParamName}} must be int")
	}
{{- end -}}
{{- template "TagTmpl" . -}}`

	TagTmpl = `
{{- $type := .Type -}}
{{- $paramname := .ParamName -}}
{{ range .Tags }}
	{{ if eq .Name ("required") }}
		if {{$paramname}} == {{if eq $type "string" }} "" {{ else }} nil {{end}} {
			return errors.New("{{$paramname}} must be not empty")
		}
	{{- end -}}
	{{- if eq .Name ("min") -}}
		{{- if eq $type ("string") -}}
			if len({{ $paramname }}) < {{.Value}} {
				return errors.New("{{$paramname}} len must be >= {{.Value}}")
			}
		{{- else -}}
			if {{$paramname}} < {{.Value}} {
				return errors.New("{{$paramname}} must be >= {{.Value}}")
			}
		{{- end -}}
	{{- end -}}
	{{- if eq .Name ("max") -}}
		{{- if eq $type ("string") -}}
			if len({{$paramname}}) > {{.Value}} {
				return errors.New("{{$paramname}} len must be <= {{.Value}}")
			}
		{{- else -}}
			if {{$paramname}} > {{.Value}} {
				return errors.New("{{$paramname}} must be <= {{.Value}}")
			}
		{{- end -}}
	{{- end -}}
	{{- if eq .Name ("enum") -}}
		if {{$paramname}} == "" {
			{{$paramname}} = "{{.DefaultValue}}"
		}
		if !(slices.Contains([]string{ {{ .QuotaValuesList }} }, {{$paramname}})) {
			return errors.New("{{$paramname}} must be one of [{{ .ValuesList }}]")
		}
	{{- end -}}
{{ end }}`

	HandleRequestTmpl = `
func (srv *{{ .BindToName }}) {{ .HandlerName }}(w http.ResponseWriter, r *http.Request) {
	{{- if .Auth }}
		if r.Header.Get("X-Auth") != "100500" {
			authErr := &ErrorApi{Error: "unauthorized"}
			response, err := json.Marshal(authErr)
			if err != nil {
				panic(err)
			}
			http.Error(w, string(response), http.StatusForbidden)
			return
		}
	{{- end }}
	params := {{.ParamsName}}{}
	err := params.Fill(r)
	if err != nil {
		paramsErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(paramsErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusBadRequest)
		return
	}
	resp, err := srv.{{ .FuncName }}(context.Background(), params)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if apiError, ok := err.(ApiError); ok {
			httpCode = apiError.HTTPStatus
		}
		handleErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), httpCode)
		return
	}
	respStruct := struct {
			Error string {{ .ErrorJSONTag }}
			Response *{{ .ResponseType }} {{ .ResponseJSONTag}} 
		}{
			Error: "",
			Response: resp,
		}
	jsonResp, err := json.Marshal(respStruct)
	if err != nil {
		panic(err)
	}
	w.Write(jsonResp)
}`

	ServeTmpl = `
func (srv *{{ .ApiName }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch  r.URL.Path {
	{{- range .Handlers -}}
		case "{{ .Path }}":
			{{- if ne .Method ("") }}
				if r.Method != "{{ .Method }}" {
					methodErr := &ErrorApi{Error: "bad method"}
					response, err := json.Marshal(methodErr)
					if err != nil {
						panic(err)
					}
					http.Error(w, string(response), http.StatusNotAcceptable)
					return
				}
			{{ end -}}
			srv.{{ .Name }}(w, r)
	{{ end -}}
	default:
		handleErr := &ErrorApi{Error: "unknown method"}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusNotFound)
	}
}`
)

var apiHandlers = make(map[string][]apiHandler)
var apiParamsCollection = make(map[string][]apiParamsField)
var apiFunctions = make(map[string][]apiFunc)

func handleStruct(currType *ast.TypeSpec) {
	currStuct, ok := currType.Type.(*ast.StructType)
	if !ok {
		return
	}

	if strings.Contains(currType.Name.Name, "Params") {
		for _, field := range currStuct.Fields.List {

			apiParamsCollection[currType.Name.Name] = append(apiParamsCollection[currType.Name.Name], apiParamsField{
				Name: field.Names[0].Name,
				Type: field.Type.(*ast.Ident).Name,
				Tag:  field.Tag.Value,
			})
		}
	}
}

func handleFunc(funcDec *ast.FuncDecl) {
	if funcDec.Doc == nil {
		return
	}

	apigenComm := ""
	comments := funcDec.Doc.List
	for _, comm := range comments {

		if strings.Contains(comm.Text, "// apigen:api") {
			apigenComm = comm.Text
		}
	}
	if apigenComm == "" {
		return
	}

	funcStruct := apiFunc{}

	fOpt := &funcOption{}
	err := json.Unmarshal([]byte(apigenComm[14:]), fOpt)
	if err != nil {
		log.Fatal(err)
	}
	funcStruct.Options = *fOpt
	funcStruct.Name = funcDec.Name.Name

	r := funcDec.Recv.List[0]
	rType, ok := r.Type.(*ast.StarExpr)
	if !ok {
		log.Fatal("Can't get func reciever")
	}
	rIdent, ok := rType.X.(*ast.Ident)
	if !ok {
		log.Fatal("Can't get func reciver")
	}
	funcStruct.BindToName = rIdent.Name

	funcType := funcDec.Type
	in := funcType.Params.List[1].Type.(*ast.Ident)
	funcStruct.InTypeName = in.Name

	outType := funcType.Results.List[0].Type
	outExp, ok := outType.(*ast.StarExpr)
	if !ok {
		log.Fatal("Can't get func out")
	}
	out, ok := outExp.X.(*ast.Ident)
	if !ok {
		log.Fatal("Can't get func out")
	}
	funcStruct.OutTypeName = out.Name

	apiFunctions[funcStruct.BindToName] = append(apiFunctions[funcStruct.BindToName], funcStruct)
}

func parseTag(tag string) map[string]string {
	res := map[string]string{}
	tagValue := strings.Split(tag, ":")[1]
	tagValue = tagValue[1 : len(tagValue)-2]
	tags := strings.Split(tagValue, ",")
	for _, t := range tags {
		keyValues := strings.Split(t, "=")
		if len(keyValues) == 1 {
			res[keyValues[0]] = ""
		} else if len(keyValues) == 2 {
			res[keyValues[0]] = keyValues[1]
		}
	}
	return res
}

func main() {
	inFileName := os.Getenv("in")
	outFileName := os.Getenv("out")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, inFileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
		return
	}
	out, err := os.Create(outFileName)
	if err != nil {
		log.Fatal(err)
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch nodeType := node.(type) {
		case *ast.File, *ast.GenDecl:
			return true
		case *ast.TypeSpec:
			handleStruct(nodeType)
			return false
		case *ast.FuncDecl:
			handleFunc(nodeType)
			return false
		default:
			return false
		}
	})

	// print package and imports
	imports := []string{"context", "encoding/json", "errors", "net/http", "slices", "strconv"}
	t, err := template.New("HeadTmpl").Parse(HeadTmpl)
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(out, struct {
		Imports      []string
		ErrorJSONTag string
	}{imports, "`json:\"error\"`"})

	// print fill and validate methods
	for name, fields := range apiParamsCollection {
		fillData := FillData{}
		fillData.Name = name
		for _, f := range fields {
			tags := parseTag(f.Tag)
			paramname := tags["paramname"]
			if paramname == "" {
				paramname = strings.ToLower(f.Name)
			}

			field := FieldData{Name: f.Name, ParamName: paramname, Type: f.Type}
			for name, value := range tags {
				switch name {
				case "min", "max":
					field.Tags = append(field.Tags, TagData{Name: name, Value: value})
				case "required":
					field.Tags = append(field.Tags, TagData{Name: name})
				case "enum":
					values := strings.Split(value, "|")
					field.Tags = append(field.Tags, TagData{
						Name:            name,
						ValuesList:      strings.Join(values, ", "),
						QuotaValuesList: `"` + strings.Join(values, `","`) + `"`,
						DefaultValue:    tags["default"],
					})
				}
			}
			fillData.Fields = append(fillData.Fields, field)
		}
		t, _ := template.New("FillTmpl").Parse(FillTmpl)
		t.New("FieldTmpl").Parse(FieldTmpl)
		t.New("TagTmpl").Parse(TagTmpl)

		t.ExecuteTemplate(out, "FillTmpl", fillData)
	}

	// print http handlers
	for api, funcs := range apiFunctions {
		for _, f := range funcs {
			funcName := fmt.Sprintf("handle%v", f.Name)
			t, err := template.New("HandleRequestTmpl").Parse(HandleRequestTmpl)
			if err != nil {
				log.Fatal(err)
			}
			t.Execute(out, HandleRequestData{
				BindToName:      f.BindToName,
				HandlerName:     funcName,
				Auth:            f.Options.Auth,
				ParamsName:      f.InTypeName,
				FuncName:        f.Name,
				ResponseType:    f.OutTypeName,
				ErrorJSONTag:    "`json:\"error\"`",
				ResponseJSONTag: "`json:\"response\"`",
			})

			apiHandlers[api] = append(apiHandlers[api], apiHandler{Path: f.Options.URL, Method: f.Options.Method, Name: funcName})
		}
	}

	// print ServeHTTP methods
	for api, handlers := range apiHandlers {
		t, err := template.New("ServeTmpl").Parse(ServeTmpl)
		if err != nil {
			log.Fatal(err)
		}
		t.Execute(out, struct {
			ApiName  string
			Handlers []apiHandler
		}{api, handlers})
	}
}
