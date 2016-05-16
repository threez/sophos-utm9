// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gorilla/mux"
	"github.com/threez/sophos-utm9/confd"
	"net/http"
)

type ObjectAPI struct {
	*SwaggerAPI
}

func (api *ObjectAPI) Get(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var value confd.AnyObject

	err := conn.GetObject("REF_"+vars["ref"], &value)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondJSON(api.MakeResty(value), w, r)
}

func (api *ObjectAPI) All(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var values []confd.AnyObject
	// TODO: transform all objects bool values to true and false

	err := conn.Request("get_objects", &values, vars["class"], vars["type"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dataArr := make([]map[string]interface{}, len(values))

	for i, value := range values {
		dataArr[i] = api.MakeResty(value)
	}

	respondJSON(dataArr, w, r)
}
