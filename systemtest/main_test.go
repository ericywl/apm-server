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

package systemtest

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("INFO: starting stack containers...")
	initContainers()
	if err := StartStackContainers(); err != nil {
		log.Fatalf("failed to start stack containers: %v", err)
	}
	initElasticSearch()
	initKibana()
	initSettings()
	initOTEL()
	log.Println("INFO: running system tests...")
	os.Exit(m.Run())
}
