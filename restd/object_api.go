// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/threez/sophos-utm9/confd"
)

// ObjectAPI handles all object related rest requests
type ObjectAPI struct {
	*SwaggerAPI
}

// RegisterAPI installs the api (CRUD) into the passed router.
// It handles requests in the /objects/{class}/{type} space.
func (api *ObjectAPI) RegisterAPI(s *mux.Router) {
	prefix := s.PathPrefix("/objects/{class}/{type}").Subrouter()

	// CREATE
	prefix.HandleFunc("/", withConfdConnection(api.PostObject)).
		Methods("POST").
		Name("createObject")

	// READ
	prefix.HandleFunc("/REF_{ref}", withConfdConnection(api.Get)).
		Methods("GET").
		Name("getObject")
	prefix.HandleFunc("/", withConfdConnection(api.GetAll)).
		Methods("GET").
		Name("getAllObject")

	// UPDATE
	prefix.HandleFunc("/REF_{ref}", withConfdConnection(api.PutObject)).
		Methods("PUT").
		Name("putObject")
	prefix.HandleFunc("/REF_{ref}", withConfdConnection(api.PatchObject)).
		Methods("PATCH").
		Name("patchObject")
	prefix.HandleFunc("/REF_{ref}", withConfdConnection(api.LockObject)).
		Methods("LOCK").
		Name("lockObject")
	prefix.HandleFunc("/REF_{ref}", withConfdConnection(api.UnlockObject)).
		Methods("UNLOCK").
		Name("unlockObject")

	// DELETE
	prefix.HandleFunc("/REF_{ref}", withConfdConnection(api.DeleteObject)).
		Methods("DELETE").
		Name("deleteObject")
}

// GetAll handles collection requests (GET /class/type)
func (api *ObjectAPI) GetAll(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var values []confd.AnyObject

	err := conn.Request("get_objects", &values, vars["class"], vars["type"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		respondJSON(err, w, r)
		return
	}
	dataArr := make([]map[string]interface{}, len(values))

	for i, value := range values {
		dataArr[i] = api.MakeResty(value)
	}

	w.WriteHeader(http.StatusOK)
	respondJSON(dataArr, w, r)
}

// PostObject handles object creation (POST /class/type)
func (api *ObjectAPI) PostObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var obj confd.AnyObject

	err := api.decodeConfdObject(&obj, r.Body, vars["class"], vars["type"])

	if err != nil {
		log.Printf("[ObjectAPI] Decoding reuqest body failed: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ref, err := conn.SetObject(obj, true)
	if err != nil {
		log.Printf("[ObjectAPI] Create %+v failed: %s", obj, err)
		w.WriteHeader(http.StatusBadRequest)
		respondJSON(err, w, r)
		return
	}
	log.Printf("[ObjectAPI] Created object %s", ref)
	w.WriteHeader(http.StatusCreated)
}

// Get handles single object requests (GET /class/type/REF_XXX)
func (api *ObjectAPI) Get(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	var value confd.AnyObject
	ref := "REF_" + vars["ref"]

	err := conn.GetObject(ref, &value)

	if err != nil {
		// if we receive an error here, most likely the REF doesn't exist.
		log.Printf("[ObjectAPI] Get %s failed: %s", ref, err)
		w.WriteHeader(http.StatusNotFound)
		respondJSON(err, w, r)
		return
	}

	if value.Class != vars["class"] || value.Type != vars["type"] {
		// there is an object, but it's from a different class, signal the error
		log.Printf("[ObjectAPI] Returning not found since wrong type %s/%s",
			value.Class, value.Type)
		w.WriteHeader(http.StatusNotFound)
		respondJSON(err, w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	vars := mux.Vars(r)
	ref := "REF_" + vars["ref"]

	_, err := conn.DelObject(ref)
	if err != nil {
		log.Printf("[ObjectAPI] Deleting %s failed: %s", ref, err)
		w.WriteHeader(http.StatusNotFound)
		respondJSON(err, w, r)
	}

	w.WriteHeader(http.StatusNoContent)
}

// LockObject sets the objects lock state to locked, then it can't be modified
// by a user or other process of the system. (LOCK /class/type/REF_XXX)
// This call doesn't validate the correctness of class and type for performance
// reasons.
func (api *ObjectAPI) LockObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	ref := "REF_" + vars["ref"]

	err := conn.LockObject(ref)
	if err != nil {
		log.Printf("[ObjectAPI] Locking %s failed: %s", ref, err)
		w.WriteHeader(http.StatusNotFound)
		respondJSON(err, w, r)
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnlockObject sets the objects lock state to unlocked, then it can be modified
// by a user or other process of the system. (UNLOCK /class/type/REF_XXX)
// This call doesn't validate the correctness of class and type for performance
// reasons.
func (api *ObjectAPI) UnlockObject(w http.ResponseWriter, r *http.Request, conn *confd.Conn) {
	vars := mux.Vars(r)
	ref := "REF_" + vars["ref"]

	err := conn.UnlockObject(ref)
	if err != nil {
		log.Printf("[ObjectAPI] Unlocking %s failed: %s", ref, err)
		w.WriteHeader(http.StatusNotFound)
		respondJSON(err, w, r)
	}

	w.WriteHeader(http.StatusNoContent)
}
