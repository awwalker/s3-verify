/*
 * s3verify (C) 2016 Minio, Inc.
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

package cmd

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	"github.com/minio/mc/pkg/console"
	"github.com/minio/mc/pkg/httptracer"
)

// traceV4 - Tracing structure for signature version '4'.
type traceV4 struct{}

// newTraceV4 - Initialize a new Trace structure.
func newTraceV4() httptracer.HTTPTracer {
	return traceV4{}
}

// Request - Trace an HTTP Request
func (t traceV4) Request(req *http.Request) (err error) {
	// Save the original Auth.
	origAuth := req.Header.Get("Authorization")

	if strings.TrimSpace(origAuth) != "" {
		// Authorization (S3 v4 signature) format:
		// Authorization: AWS4-HMAC-SHA256 Credential=AKIAJNACEGBGMXBHLEZA/20150524/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=bbfaa693c626021bcb5f911cd898a1a30206c1fad6bad1e0eb89e282173bd24c

		// Strip out accessKeyID from: Credential=<access-key-id>/<date>/<aws-region>/<aws-service><aws4_request
		regCred := regexp.MustCompile("Credential=([A-Z0-9]+)/")
		redactedAuth := regCred.ReplaceAllString(origAuth, "Credential=**REDACTED**/")

		// Strip out 256-bit signature from: Signature=<256-bit signature>
		regSign := regexp.MustCompile("Signature=([0-9a-f]+)")
		redactedAuth = regSign.ReplaceAllString(redactedAuth, "Signature=**REDACTED**")

		// Use a temporary header.
		req.Header.Set("Authorization", redactedAuth)

		var reqTrace []byte
		reqTrace, err = httputil.DumpRequestOut(req, false) // Only display header, no body.
		if err == nil {
			console.Eraseline() // Remove the scanBar line.
			console.Debug(string(reqTrace))
		}

		// Replace with real header.
		req.Header.Set("Authorization", origAuth)
	}
	return err
}

func (t traceV4) Response(res *http.Response) (err error) {
	var respTrace []byte
	// For errors dump response body as well.
	if res.StatusCode != http.StatusOK &&
		res.StatusCode != http.StatusPartialContent &&
		res.StatusCode != http.StatusNoContent {
		respTrace, err = httputil.DumpResponse(res, true)
	} else {
		if res.ContentLength == 0 {
			var buffer bytes.Buffer
			if err = res.Header.Write(&buffer); err != nil {
				return err
			}
			respTrace = buffer.Bytes()
			respTrace = append(respTrace, []byte("\r\n")...)
		} else {
			respTrace, err = httputil.DumpResponse(res, false)
			if err != nil {
				return err
			}
		}
	}
	if err == nil {
		console.Eraseline() // Remove the scanBar line.
		console.Debug(string(respTrace))
	}
	return err
}
