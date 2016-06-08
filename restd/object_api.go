// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gorilla/mux"
	"github.com/threez/sophos-utm9/confd"
	"net/http"
)

// ObjectAPI handles all object related rest requests
type ObjectAPI struct {
	*SwaggerAPI
}

// RegisterAPI installs the api (CRUD) into the passed router.
// It handles requests in the /objects/{class}/{type} space.
func (api *ObjectAPI) RegisterAPI(s *mux.Router) {
	objAPI := s.PathPrefix("/objects/{class}/{type}")

	// CREATE
	create := objAPI.Methods("POST").Subrouter()
	create.HandleFunc("/", withConfdConnection(api.PostObject)).Name("createObject")

	// READ
	read := objAPI.Methods("GET").Subrouter()
	read.HandleFunc("/REF_{ref}", withConfdConnection(api.Get)).Name("getObject")
	read.HandleFunc("/", withConfdConnection(api.GetAll)).Name("getAllObject")

	// UPDATE
	objAPI.Methods("PUT").Subrouter().
		HandleFunc("/REF_{ref}", withConfdConnection(api.PutObject)).Name("putObject")
	objAPI.Methods("PATCH").Subrouter().
		HandleFunc("/REF_{ref}", withConfdConnection(api.PatchObject)).Name("patchObject")
	objAPI.Methods("LOCK").Subrouter().
		HandleFunc("/REF_{ref}", withConfdConnection(api.LockObject)).Name("lockObject")
	objAPI.Methods("UNLOCK").Subrouter().
		HandleFunc("/REF_{ref}", withConfdConnection(api.UnlockObject)).Name("unlockObject")

	// DELETE
	delete := objAPI.Methods("DELETE").Subrouter()
	delete.HandleFunc("/REF_{ref}", withConfdConnection(api.DeleteObject)).Name("deleteObject")
}

// GetAll handles collection requests (GET /class/type)
func (api *ObjectAPI) GetAll(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var values []confd.AnyObject

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

// PostObject handles object creation (POST /class/type)
func (api *ObjectAPI) PostObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get handles single object requests (GET /class/type/REF_XXX)
func (api *ObjectAPI) Get(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var value confd.AnyObject

	err := conn.GetObject("REF_"+vars["ref"], &value)

	if err != nil {
		// if we receive an error here, most likely the REF doesn't exist.
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if value.Class != vars["class"] || value.Type != vars["type"] {
		// there is an object, but it's from a different class, signal the error
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respondJSON(api.MakeResty(value), w, r)
}

// PutObject handles single object update requests (PUT /class/type/REF_XXX).
// Requires the full object to be send. If not all parameters are passed,
// will return BadRequest.
func (api *ObjectAPI) PutObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PatchObject handles single object change requests (PATCH /class/type/REF_XXX)
// This call doesn't validate the correctness of class and type for performance
// reasons.
func (api *ObjectAPI) PatchObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	w.WriteHeader(http.StatusNotImplemented)
}

// DeleteObject handles single object deletion requests
// (DELETE /class/type/REF_XXX)
// This call doesn't validate the correctness of class and type for performance
// reasons.
func (api *ObjectAPI) DeleteObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	w.WriteHeader(http.StatusNotImplemented)
}

// LockObject sets the objects lock state to locked, then it can't be modified
// by a user or other process of the system. (LOCK /class/type/REF_XXX)
// This call doesn't validate the correctness of class and type for performance
// reasons.
func (api *ObjectAPI) LockObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	w.WriteHeader(http.StatusNotImplemented)
}

// UnlockObject sets the objects lock state to unlocked, then it can be modified
// by a user or other process of the system. (UNLOCK /class/type/REF_XXX)
// This call doesn't validate the correctness of class and type for performance
// reasons.
func (api *ObjectAPI) UnlockObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	w.WriteHeader(http.StatusNotImplemented)
}
