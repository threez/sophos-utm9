// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/threez/sophos-utm9/confd"
)

type SwaggerInfo struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

type SwaggerPaths map[string]*SwaggerAction

type Items map[string]string

type SwaggerSchema struct {
	Type        string                    `json:"type,omitempty"`
	Format      string                    `json:"format,omitempty"`
	Description string                    `json:"description,omitempty"`
	Properties  map[string]*SwaggerSchema `json:"properties,omitempty"`
	Items       Items                     `json:"items,omitempty"`
	Ref         string                    `json:"$ref,omitempty"`
	Default     interface{}               `json:"default,omitempty"`
}

type SwaggerResponse struct {
	Description string         `json:"description,omitempty"`
	Schema      *SwaggerSchema `json:"schema,omitempty"`
}

type SwaggerParameter struct {
	Name        string         `json:"name,omitempty"`
	In          string         `json:"in,omitempty"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Type        string         `json:"type,omitempty"`
	Schema      *SwaggerSchema `json:"schema,omitempty"`
}

type SwaggerAction struct {
	Summary     string                      `json:"summary,omitempty"`
	Description string                      `json:"description,omitempty"`
	Parameters  []*SwaggerParameter         `json:"parameters,omitempty"`
	Tags        []string                    `json:"tags,omitempty"`
	Responses   map[string]*SwaggerResponse `json:"responses,omitempty"`
}

type BasicAuth struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	In   string `json:"in,omitempty"`
}

type Swagger struct {
	SpecVersion         string                    `json:"swagger,omitempty"`
	Info                SwaggerInfo               `json:"info,omitempty"`
	Host                string                    `json:"host,omitempty"`
	Schemes             []string                  `json:"schemes,omitempty"`
	BasePath            string                    `json:"basePath,omitempty"`
	Produces            []string                  `json:"produces,omitempty"`
	Paths               map[string]SwaggerPaths   `json:"paths,omitempty"`
	Definitions         map[string]*SwaggerSchema `json:"definitions,omitempty"`
	SecurityDefinitions map[string]*BasicAuth     `json:"securityDefinitions,omitempty"`
}

var translation = map[string]string{
	"ARRAY":      "array",
	"BLOB":       "string",
	"BOOL":       "boolean",
	"DATE":       "string",
	"EMAIL":      "string",
	"HASH":       "string",
	"HEXSTRING":  "string",
	"HOSTNAME":   "string",
	"INTEGER":    "integer",
	"IP6ADDR":    "string",
	"IPADDR":     "string",
	"LISTPICK":   "string",
	"MACADDR":    "string",
	"REF":        "string",
	"REGEX":      "string",
	"SNMPSTRING": "string",
	"TIME":       "string",
}

var translationFormat = map[string]string{
	"BLOB": "byte",
	"DATE": "date",
	"TIME": "date-time",
}

type SwaggerAPI struct {
	Classes []string
	Meta    confd.ObjectMetaTree
	Specs   map[string]*Swagger
	conn    *confd.Conn
	prefix  string
}

func NewSwaggerAPI(apiPrefix string) (*SwaggerAPI, error) {
	var err error
	api := &SwaggerAPI{
		conn:   confd.NewSystemConn(),
		prefix: apiPrefix,
	}
	api.conn.Logger = confdLogger
	// after querying the meta data, the connection can be closed
	defer func() { _ = api.conn.Close() }()

	// fetch object metadata
	api.Meta, err = api.conn.GetMetaObjects()
	if err != nil {
		return nil, err
	}

	// build object scwagger specs
	api.Specs, err = api.prebuildSwaggerSpecs()
	return api, err
}

func (a *SwaggerAPI) RegisterSwaggerAPI(r *mux.Router) {
	type ClassLink struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Link        string `json:"link"`
	}

	r.HandleFunc("/classes", func(w http.ResponseWriter, r *http.Request) {
		classDefs := make([]ClassLink, len(a.Classes))
		for i, class := range a.Classes {
			classDefs[i].Description = class
			classDefs[i].Link = fmt.Sprintf(a.prefix+"/%s/swagger.json", class)
			classDefs[i].Name = class
		}
		respondJSON(classDefs, w, r)
	})

	r.HandleFunc("/{class}/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		var class = mux.Vars(r)["class"]
		respondJSON(a.Specs[class], w, r)
	})

	r.PathPrefix("/").Handler(
		http.StripPrefix(a.prefix, http.FileServer(http.Dir("./static/"))))
}

func (a *SwaggerAPI) Cors(handler http.Handler) http.Handler {
	c := cors.New(cors.Options{
		// AllowedOrigins is a list of origins a cross-domain request can be executed from.
		// If the special "*" value is present in the list, all origins will be allowed.
		// An origin may contain a wildcard (*) to replace 0 or more characters
		// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penality.
		// Only one wildcard can be used per origin.
		// Default value is ["*"]
		AllowedOrigins: []string{"*"}, //http://localhost:3000
		// AllowOriginFunc is a custom function to validate the origin. It take the origin
		// as argument and returns true if allowed or false otherwise. If this option is
		// set, the content of AllowedOrigins is ignored.
		AllowOriginFunc: nil,
		// AllowedMethods is a list of methods the client is allowed to use with
		// cross-domain requests. Default value is simple methods (GET and POST)
		AllowedMethods: []string{"GET", "PUT", "OPTIONS", "POST", "DELETE", "PATCH", "LOCK", "UNLOCK"},
		// AllowedHeaders is list of non simple headers the client is allowed to use with
		// cross-domain requests.
		// If the special "*" value is present in the list, all headers will be allowed.
		// Default value is [] but "Origin" is always appended to the list.
		AllowedHeaders: []string{"Authorization"},
		// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
		// API specification
		ExposedHeaders: []string{},
		// AllowCredentials indicates whether the request can include user credentials like
		// cookies, HTTP authentication or client side SSL certificates.
		AllowCredentials: true,
		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached
		MaxAge: 0,
		// OptionsPassthrough instructs preflight to let other potential next handlers to
		// process the OPTIONS method. Turn this on if your application handles OPTIONS.
		OptionsPassthrough: false,
		// Debugging flag adds additional output to debug server side CORS issues
		Debug: false,
	})
	return c.Handler(handler)
}

func (a *SwaggerAPI) MakeResty(obj confd.AnyObject) map[string]interface{} {
	data := obj.Data

	// make bool values more friendly to the user
	for key, value := range data {
		definition := a.Meta[obj.Class][obj.Type][key]
		if definition.ISA == "BOOL" {
			data[key] = (value == 1)
		}
	}

	data["_ref"] = obj.Ref
	data["_type"] = fmt.Sprintf("%s/%s", obj.Class, obj.Type)
	data["_locked"] = bool(obj.Lock)

	return data
}

// prebuildSwaggerSpecs returns the swagger specs for all classes
func (a *SwaggerAPI) prebuildSwaggerSpecs() (map[string]*Swagger, error) {
	var classes []string
	err := a.conn.Request("get_object_classes", &classes)
	if err != nil {
		return nil, err
	}
	sort.Strings(classes)
	a.Classes = classes

	schemas := make(map[string]*Swagger)
	for _, class := range classes {
		log.Printf("Generating swagger spec for %s", class)
		schemas[class], err = a.buildSwagger(class)
		if err != nil {
			return nil, err
		}
	}
	return schemas, nil
}

func (a *SwaggerAPI) buildSwagger(class string) (*Swagger, error) {
	responses := make(map[string]*SwaggerResponse)
	responses["200"] = &SwaggerResponse{Description: "Ok"}
	responses["404"] = &SwaggerResponse{Description: "NotFound"}

	paths := make(map[string]SwaggerPaths)

	var classTypes []string
	err := a.conn.Request("get_object_types", &classTypes, class)
	// TODO: handle error
	for _, classType := range classTypes {
		tpath := fmt.Sprintf("/objects/%s/%s", class, classType)
		objectResponses := make(map[string]*SwaggerResponse)
		objectResponses["200"] = &SwaggerResponse{
			Description: fmt.Sprintf("%s::%s objects", class, classType),
			Schema: &SwaggerSchema{
				Type: "array",
				Items: map[string]string{
					"$ref": fmt.Sprintf("#/definitions/%s::%s", class, classType),
				},
			},
		}

		paths[tpath] = make(map[string]*SwaggerAction)
		paths[tpath]["get"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s::%s type", class, classType),
			Description: fmt.Sprintf("Returns all available %s::%s objects", class, classType),
			Parameters:  []*SwaggerParameter{},
			Tags:        []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses:   objectResponses,
		}

		paths[tpath]["post"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s::%s type", class, classType),
			Description: fmt.Sprintf("Create a new %s::%s object", class, classType),
			Parameters: []*SwaggerParameter{
				&SwaggerParameter{
					In:          "body",
					Name:        "body",
					Description: fmt.Sprintf("%s::%s that will be created", class, classType),
					Required:    true,
					Schema: &SwaggerSchema{
						Ref: fmt.Sprintf("#/definitions/%s::%s", class, classType),
					},
				},
			},
			Tags:      []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses: responses,
		}
		pathSingle := fmt.Sprintf("/objects/%s/%s/{ref}", class, classType)
		paths[pathSingle] = make(map[string]*SwaggerAction)

		refParameters := []*SwaggerParameter{
			&SwaggerParameter{
				Name:        "ref",
				In:          "path",
				Description: "id of the object",
				Required:    true,
				Type:        "string",
			},
		}

		paths[pathSingle]["get"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s class", class),
			Description: fmt.Sprintf("Returns all available %s types", classType),
			Parameters:  refParameters,
			Tags:        []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses:   responses,
		}
		paths[pathSingle]["put"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s class", class),
			Description: fmt.Sprintf("Creates or updates the complete object %s", classType),
			Parameters: []*SwaggerParameter{
				&SwaggerParameter{
					In:          "body",
					Name:        "body",
					Description: fmt.Sprintf("%s::%s that will be updated", class, classType),
					Required:    true,
					Schema: &SwaggerSchema{
						Ref: fmt.Sprintf("#/definitions/%s::%s", class, classType),
					},
				},
			},
			Tags:      []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses: responses,
		}
		paths[pathSingle]["delete"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s class", class),
			Description: fmt.Sprintf("Creates or updates the complete object %s", classType),
			Parameters:  refParameters,
			Tags:        []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses:   responses,
		}
		paths[pathSingle]["patch"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s class", class),
			Description: fmt.Sprintf("Changes to parts of the object %s types", classType),
			Parameters: []*SwaggerParameter{
				&SwaggerParameter{
					In:          "body",
					Name:        "body",
					Description: fmt.Sprintf("%s::%s that will be changes", class, classType),
					Required:    true,
					Schema: &SwaggerSchema{
						Ref: fmt.Sprintf("#/definitions/%s::%s", class, classType),
					},
				},
			},
			Tags:      []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses: responses,
		}
		paths[pathSingle]["lock"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s class", class),
			Description: fmt.Sprintf("Locks the object %s types", classType),
			Parameters:  refParameters,
			Tags:        []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses:   responses,
		}
		paths[pathSingle]["unlock"] = &SwaggerAction{
			Summary:     fmt.Sprintf("%s class", class),
			Description: fmt.Sprintf("Unlocks the object %s types", classType),
			Parameters:  refParameters,
			Tags:        []string{fmt.Sprintf("%s::%s", class, classType)},
			Responses:   responses,
		}
	}

	definitions := make(map[string]*SwaggerSchema)
	classmeta := a.Meta[class]
	for _, classType := range classTypes {
		attrs := classmeta[classType]
		name := fmt.Sprintf("%s::%s", class, classType)
		var desc string
		a.conn.Request("get_object_descr", &desc, class, classType)
		schema := &SwaggerSchema{
			Type:        "object",
			Description: desc,
			Properties:  make(map[string]*SwaggerSchema),
		}
		// TODO: handle error
		for attr, def := range attrs {
			if attr == "_name" {
				continue
			}
			schema.Properties[attr] = mapProperty(def)
		}
		schema.Properties["name"] = &SwaggerSchema{Type: "string"}
		schema.Properties["comment"] = &SwaggerSchema{Type: "string"}
		definitions[name] = schema
	}

	return &Swagger{
		SpecVersion: "2.0",
		Info: SwaggerInfo{
			Title:       "SOPHOS UTM9 REST API",
			Description: "Restful sophos confd API",
			Version:     "1.0.0",
		},
		Schemes:     []string{"http"},
		BasePath:    a.prefix,
		Produces:    []string{"application/json"},
		Paths:       paths,
		Definitions: definitions,
		SecurityDefinitions: map[string]*BasicAuth{
			"auth": &BasicAuth{
				Type: "basic",
				Name: "basicAuth",
				In:   "header",
			},
		},
	}, err
}

func mapProperty(field confd.AttrConstraintWrapper) *SwaggerSchema {
	t := translation[field.ISA]
	if t == "" {
		t = "string"
	}

	if t == "array" {
		items := map[string]string{
			"type": translation[field.Type],
		}
		if format := translationFormat[field.Type]; format != "" {
			items["format"] = format
		}

		return &SwaggerSchema{
			Type:  t,
			Items: items,
		}
	}

	return &SwaggerSchema{
		Type:    t,
		Format:  translationFormat[field.Type],
		Default: field.Default,
	}
}
