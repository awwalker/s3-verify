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
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// overWrittenHeaers - map the request headers that can be sent
// to the response headers they will overwrite.
// These are the headers that support overwriting according to
// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectGET.html
var overWrittenHeaders = map[string]string{
	"response-content-type":        "Content-Type",
	"response-content-language":    "Content-Language",
	"response-expires":             "Expires",
	"response-cache-control":       "Cache-Control",
	"response-content-disposition": "Content-Disposition",
	"response-content-encoding":    "Content-Encoding",
}

// newGetObjectReq - Create a new HTTP requests to perform.
func newGetObjectReq(bucketName, objectName string, responseHeaders map[string]string) (Request, error) {
	// getObjectReq - a new HTTP request for a GET object.
	var getObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	getObjectReq.bucketName = bucketName
	getObjectReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	urlValues := make(url.Values)
	// Overide response headers.
	for k, v := range responseHeaders {
		urlValues.Set(k, v)
	}
	getObjectReq.queryValues = urlValues

	// Set the headers.
	getObjectReq.customHeader.Set("User-Agent", appUserAgent)
	getObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return getObjectReq, nil
}

// TODO: These checks only verify correctly formatted requests. There is no request that is made to fail / check failure yet.

// getObjectVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func getObjectVerify(res *http.Response, expectedBody []byte, expectedStatusCode int, expectedHeader map[string]string) error {
	if err := verifyHeaderGetObject(res.Header, expectedHeader); err != nil {
		return err
	}
	if err := verifyStatusGetObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyGetObject(res.Body, expectedBody); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObject - Verify that the header returned matches what is expected.
func verifyHeaderGetObject(header http.Header, expectedHeaders map[string]string) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	for k, v := range expectedHeaders {
		headerKey := overWrittenHeaders[k]
		if header.Get(headerKey) != v {
			err := fmt.Errorf("Unexpected Header Value Received for %s: wanted %s, got %s", k, v, header.Get(headerKey))
			return err
		}
	}
	return nil
}

// verifyBodyGetObject - Verify that the body returned matches what is expected.
func verifyBodyGetObject(resBody io.Reader, expectedBody []byte) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// Compare what was created to be uploaded and what is contained in the response body.
	if !bytes.Equal(body, expectedBody) {
		err := fmt.Errorf("Unexpected Body Recieved: wanted %v, got %v", string(expectedBody), string(body))
		return err
	}
	return nil
}

// verifyStatusGetObject - Verify that the status returned matches what is expected.
func verifyStatusGetObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// mainGetObject - test a get object request.
func mainGetObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject:", curTest, globalTotalNumTest)
	// Use the bucket created in the mainPutBucketPrepared Test.
	// Set the response headers to be overwritten.
	expectedHeaders := map[string]string{
		"response-content-type": "image/gif",
		//		"response-content-language":    "da",
		"response-expires":             "Thu, 01 Dec 1994 16:00:00 GMT",
		"response-cache-control":       "no-cache",
		"response-content-disposition": "attachment; filename=\"s3verify.txt\"",
		//		"response-content-encoding":    "gzip",
	}
	// All getobject tests happen in s3verify created buckets
	// on s3verify objects.
	bucketName := s3verifyBuckets[0].Name
	for _, object := range s3verifyObjects {
		// Spin scanBar
		scanBar(message)
		// Create new GET object request.
		req, err := newGetObjectReq(bucketName, object.Key, expectedHeaders)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("GET", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := getObjectVerify(res, object.Body, http.StatusOK, expectedHeaders); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
