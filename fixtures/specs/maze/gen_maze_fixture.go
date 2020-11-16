// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type KeyRefable struct {
	Key  string
	Refs []string
}

type NameRefable struct {
	Name string
	Ref  string
}

const (
	SchemaTypePlain   = "plain"
	SchemaTypePropRef = "prop-ref"
	SchemaTypeItemRef = "items-ref"

	RefTypeLocal     = "local"
	RefTypeRelative  = "relative"
	RefTypeAbsFile   = "abs-file"
	RefTypeAbsRemote = "abs-remote"
)

type GenParameter struct {
	In     string
	Name   string
	Schema GenSchema
}

type GenCommonParameter struct {
	KeyRefable
	Parameter GenParameter
}

type GenCommonResponse struct {
	KeyRefable
	Response GenResponse
}

type GenResponse struct {
	Code   string
	Ref    string
	Schema GenSchema
}

type GenSchema struct {
	NameRefable
	SchemaType string
}

type GenPath struct {
	KeyRefable
	Parameters           []GenParameter
	OperationsParameters []GenParameter
	Responses            []GenResponse
	CommonParameters     []GenCommonParameter // spec-level parameters
	CommonResponses      []GenCommonResponse  // spec-level responses
}

type GenSpec struct {
	Description      string
	Schemas          []GenSchema
	CommonParameters []GenCommonParameter // spec-level parameters
	CommonResponses  []GenCommonResponse  // spec-level responses
	Paths            []GenPath
}

type GenConfig struct {
	BaseDir        string
	BaseSchemaName string
	RemoteHost     string
	RemotePort     int
}

func main() {
	// this generates a large swagger spec with a focus on systematically exploring $ref combinations
	pwd, _ := os.Getwd()
	config := GenConfig{
		BaseDir:    pwd,
		RemoteHost: "localhost",
		RemotePort: 1234,
	}
	funcMap := makeFuncMap(config)
	spec := GenSpec{
		Description:      `A generated spec to explore $ref's`,
		CommonParameters: makeCommonParameters(config),
		CommonResponses:  makeCommonResponses(config),
		Schemas:          makeSchemas(config), // definitions at spec level
		Paths:            makePaths(config),
	}
	templateText, err := ioutil.ReadFile("templates/spec.gotmpl")
	if err != nil {
		log.Fatalf("loading: %s", err)
	}
	tmpl, err := template.New("genSpec").Funcs(funcMap).Parse(string(templateText))
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}
	err = tmpl.Execute(os.Stdout, spec)
	if err != nil {
		log.Fatalf("execution: %s", err)
	}
	// TODO: parameters/parameters.yaml
	// TODO: responses/responses.yaml
	// TODO: paths/paths.yaml
	// TODO: schemas/definitions.yaml
}

func makeParamRefs(config GenConfig, radix string) []string {
	// in addition to '#/parameters',
	// expects ref resolved in:
	// - http://{host:port}/remote/parameters.yaml
	// - file://{PWD}/parameters/parameters.yaml
	// - parameters/parameters.yaml
	return []string{
		fmt.Sprintf("http://%s:%d", config.RemoteHost, config.RemotePort) +
			"/" + path.Join("remote", "parameters.yaml#", fmt.Sprintf("local-remote-abs-%s", radix)), // absolute remote $ref
		fmt.Sprintf("file://%s",
			filepath.ToSlash(filepath.Join(config.BaseDir, "parameters", "parameters.yaml#", fmt.Sprintf("local-file-abs-%s", radix)))), // absolue file $ref
		path.Join("parameters", "parameters.yaml#", fmt.Sprintf("local-file-rel-%s", radix)), // relative file $ref
	}
}

func makeSchemaRef(config GenConfig, definition, refType, schemaType string) string {
	if SchemaType == SchemaTypePlain {
		return ""
	}
	switch refType {
	case RefTypeLocal:
		return path.Join("#", "definitions", definition)
	case RefTypeRelative:
		return path.Join("schemas", "definitions.yaml#", definition)
	case RefTypeAbsFile:
		return fmt.Sprintf("file://%s",
			filepath.ToSlash(filepath.Join(config.BaseDir, "definitions", "definitions.yaml#", definition)))
	case RefTypeAbsRemote:
		return fmt.Sprintf("http://%s:%d", config.RemoteHost, config.RemotePort) +
			"/" + path.Join("remote", "definitions.yaml#", definition)
	default:
		panic(fmt.Sprintf("invalid ref type: %s", refType))
	}
}

func makeCommonParameters(config GenConfig) []GenCommonParameter {
	result := make([]GenCommonParameter, 0, 24)
	for _, radix := range []string{"parameter-body", "parameter-query"} {
		in := strings.Split(radix, "-")[1]
		for _, schemaType := range []string{SchemaTypePlain, SchemaTypePropRef, SchemaTypeItemRef} {
			for _, refType := range []string{RefTypeLocal, RefTypeRelative, RefTypeAbsFile, RefTypeAbsRemote} {

				key := fmt.Sprintf("local-%s-%s-%s", schemaType, refType, radix)
				name := key
				definition := fmt.Sprintf("local-def-%s-%s-%s", schemaType, refType, radix)

				result = append(result,
					GenCommonParameter{
						KeyRefable: KeyRefable{
							Key:  key,                          // the key in "parameters" section
							Refs: makeParamRefs(config, radix), // all other refs to generate from this base parameter
						},
						GenParameter{
							In:   in,
							Name: name,
							Schema: GenSchema{
								SchemaType: refType,
								NameRefable: NameRefable{
									Name: definition,                                             // name in definition
									Ref:  makeSchemaRef(config, definition, refType, SchemaType), // embedded ref
								},
							},
						},
					})
			}
		}
	}
	return result
}

func makeResponseCode(i int) string {
	if i == 0 {
		return "default"
	}
	return strconv.Itoa((int(i/20)+2)*100 + i%20)
}

func makeCommonResponses(config GenConfig) []GenCommonResponse {
	result := make([]GenCommonResponse, 0, 12)
	radix := "resp"
	index := 0
	for _, schemaType := range []string{SchemaTypePlain, SchemaTypePropRef, SchemaTypeItemRef} {
		for _, refType := range []string{RefTypeLocal, RefTypeRelative, RefTypeAbsFile, RefTypeAbsRemote} {
			key := fmt.Sprintf("local-%s-%s-%s", schemaType, refType, radix)
			definition := fmt.Sprintf("local-def-%s-%s-%s", schemaType, refType, radix)
			result = append(result,
				GenCommonResponse{
					KeyRefable: KeyRefable{
						Key:  key,                          // the key in "parameters" section
						Refs: makeParamRefs(config, radix), // all other refs to generate from this base
					},
					Response: GenResponse{
						Code: makeResponseCode(index),
						Ref:  "", // common responses are not themselves made of a $ref
						Schema: GenSchema{
							SchemaType: SchemaType,
							NameRefable: NameRefable{
								Name: definition,                                             // name in definition
								Ref:  makeSchemaRef(config, definition, refType, schemaType), // embedded ref
							},
						},
					},
				})
			index++
		}
	}
	return result
}

func makeSchemas(config GenConfig) []GenSchema {
	return nil
}

func makePaths(config GenConfig) []GenPath {
	return nil
}

func makeFuncMap() template.FuncMap {
	return template.FuncMap{
		"indent": func(text string, indent int) string {
			return ""
		},
		"refName": func(ref string) string {
			return ""
		},
	}
}
