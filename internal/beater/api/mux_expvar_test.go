// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/apm-server/internal/beater/config"
)

func TestExpvarDefaultDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	recorder, err := requestToMuxerWithPattern(t, cfg, "/debug/vars")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, `{"error":"404 page not found"}`+"\n", recorder.Body.String())
}

func TestExpvarEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Expvar.Enabled = true
	recorder, err := requestToMuxerWithPattern(t, cfg, "/debug/vars")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)

	decoded := make(map[string]interface{})
	err = json.NewDecoder(recorder.Body).Decode(&decoded)
	assert.NoError(t, err)
	assert.Contains(t, decoded, "memstats")
}
