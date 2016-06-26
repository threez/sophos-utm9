// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
)

func respondJSON(values interface{}, w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(values, "", "  ")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(bytes) // ignore error, as there is no meaningful handling
}
