/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

var (
	globalVerbose       = false
	globalDefaultRegion = "us-east-1"
	globalTotalNumTest  = 0
	globalRandom        *rand.Rand
)

// lockedRandSource provides protected rand source, implements rand.Source interface.
type lockedRandSource struct {
	lk  sync.Mutex
	src rand.Source
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an
// int64.
func (r *lockedRandSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

// Seed uses the provided seed value to initialize the generator to a
// deterministic state.
func (r *lockedRandSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

// Separate out context.
func setGlobals(verbose bool, numTests int) {
	globalTotalNumTest = numTests
	globalVerbose = verbose
	if globalVerbose {
		// Allow printing of traces.
		console.DebugPrint = true
	}
	globalRandom = rand.New(&lockedRandSource{src: rand.NewSource(time.Now().UTC().UnixNano())})
}

// Set any global flags here.
func setGlobalsFromContext(ctx *cli.Context) error {
	verbose := ctx.Bool("verbose") || ctx.GlobalBool("verbose")
	numTests := 0
	if ctx.Bool("extended") || ctx.GlobalBool("extended") {
		numTests = len(unpreparedTests)
	} else {
		// The length of unpreparedTests == preparedTests.
		for _, test := range unpreparedTests {
			if !test.Extended {
				numTests++
			}
		}
	}
	setGlobals(verbose, numTests)

	return nil
}
