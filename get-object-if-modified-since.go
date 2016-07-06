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
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/s3verify/signv4"
)

// GetObjectIfModifiedReq - an HTTP GET request with the If-Modified-Since header set.
var GetObjectIfModifiedReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for GET requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		"If-Modified-Since":    {""}, // To be added dynamically.
	},
	Body:   nil, // There is no body for GET requests.
	Method: "GET",
}

// NewGetObjcetIfModifiedSinceReq - Create a new HTTP request to perform.
func NewGetObjectIfModifiedSinceReq(config ServerConfig, bucketName, objectName, lastModified string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	GetObjectIfModifiedReq.Header["If-Modified-Since"] = []string{lastModified}

	// Fill request URL and sign.
	GetObjectIfModifiedReq.URL = targetURL
	GetObjectIfModifiedReq = signv4.SignV4(*GetObjectIfModifiedReq, config.Access, config.Secret, config.Region)
	return GetObjectIfModifiedReq, nil
}

// GetObjectIfModifiedSinceInit - Set up a new bucket and object to perform the request on.
func GetObjectIfModifiedSinceInit(s3Client minio.Client, config ServerConfig) (bucketName, objectName, lastModified string, buf []byte, err error) {
	// Create a new random bucket and object name prefixed by s3verify-get.
	bucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	objectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	lastModified = ""
	// Create random data more than 32K.
	buf = make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Create the test bucket and object.
	err = s3Client.MakeBucket(bucketName, config.Region)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Upload the random object.
	_, err = s3Client.PutObject(bucketName, objectName, bytes.NewReader(buf), "binary/octet-stream")
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Gather the Last-Modified field of the object.
	objInfo, err := s3Client.StatObject(bucketName, objectName)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	lastModifiedTime := objInfo.LastModified
	lastModified = lastModifiedTime.Format(http.TimeFormat)
	return bucketName, objectName, lastModified, buf, err
}

// VerifyGetObjectIfModifiedSince - Verify that the response matches what is expected.
func VerifyGetObjectIfModifiedSince(res *http.Response, expectedBody []byte, expectedStatus string, expectedHeader map[string][]string) error {
	if err := VerifyHeaderGetObjectIfModifiedSince(res); err != nil {
		return err
	}
	if err := VerifyBodyGetObjectIfModifiedSince(res, expectedBody); err != nil {
		return err
	}
	if err := VerifyStatusGetObjectIfModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// VerifyBodyGetObjectIfModifiedSince - Verify that the response body matches what is expected.
func VerifyBodyGetObjectIfModifiedSince(res *http.Response, expectedBody []byte) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, expectedBody) {
		err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(expectedBody), string(body))
		return err
	}
	return nil
}

// VerifyStatusGetObjectIfModifiedSince - Verify that the response status matches what is expected.
func VerifyStatusGetObjectIfModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyHeaderGetObjectIfModifiedSince - Verify that the response header matches what is expected.
func VerifyHeaderGetObjectIfModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test the compatibility of the GET object API when using the If-Modified-Since header.
func mainGetObjectIfModifiedSince(config ServerConfig, s3Client minio.Client, message string) error {
	// Set a date in the past.
	pastDate := "Thu, 01 Jan 1970 00:00:00 GMT"
	// Spin scanBar
	scanBar(message)
	// Set up a new bucket and object to GET against.
	bucketName, objectName, lastModified, buf, err := GetObjectIfModifiedSinceInit(s3Client, config)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Create new GET object request.
	req, err := NewGetObjectIfModifiedSinceReq(config, bucketName, objectName, lastModified)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response...these checks do not check the headers yet.
	if err := VerifyGetObjectIfModifiedSince(res, []byte(""), "304 Not Modified", nil); err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Create an acceptable request.
	goodReq, err := NewGetObjectIfModifiedSinceReq(config, bucketName, objectName, pastDate)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute the response that should give back a body.
	goodRes, err := ExecRequest(goodReq, config.Client)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify that the past date gives back the data.
	if err := VerifyGetObjectIfModifiedSince(goodRes, buf, "200 OK", nil); err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := CleanUpGetObject(s3Client, bucketName, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Clean up after the test
	if err := CleanUpGetObject(s3Client, bucketName, []string{objectName}); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	return nil
}
