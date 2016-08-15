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
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go"
)

const numTestObjects = 101

// prepareBucket - Uses minio-go library to create new testing bucket for use by s3verify.
func prepareBuckets(region string, client *minio.Client) (string, error) {
	message := "Creating test bucket"
	bucketName := "s3verify-" + globalSuffix
	// Spin scanBar
	scanBar(message)
	err := client.MakeBucket(bucketName, region)
	if err != nil {
		printMessage(message, err)
		return "", err
	}
	// Spin scanBar
	scanBar(message)
	// Bucket preparation passed.
	printMessage(message, nil)
	return bucketName, nil
}

// TODO: see if parallelization has a place here.

// prepareObjects - Uses minio-go library to create 1001 new testing objects for use by s3verify.
func prepareObjects(client *minio.Client, bucketName string) error {
	message := "Creating test objects"
	// TODO: update this to 1001...for testing purposes it is OK to leave it at 101 for now.
	// Upload 1001 objects specifically for the list-objects tests.
	for i := 0; i < numTestObjects; i++ {
		// Spin scanBar
		scanBar(message)
		randomData := randString(60, rand.NewSource(time.Now().UnixNano()), "")
		objectKey := "s3verify/put/object/" + globalSuffix + strconv.Itoa(i)
		// Create 60 bytes worth of random data for each object.
		reader := bytes.NewReader([]byte(randomData))
		_, err := client.PutObject(bucketName, objectKey, reader, "application/octet-stream")
		if err != nil {
			printMessage(message, err)
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	randomData := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	objectKey := "s3verify/list/" + globalSuffix
	reader := bytes.NewReader([]byte(randomData))
	_, err := client.PutObject(bucketName, objectKey, reader, "application/octet-stream")
	if err != nil {
		printMessage(message, err)
	}
	// Object preparation passed.
	printMessage(message, nil)
	return nil
}

// validateBucket - validates that the bucket passed to s3verify was created by s3verify.
func validateBucket(config ServerConfig, bucketName string) error {
	// Create a new minio-go client object.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	secure := false
	if hostURL.Scheme == "https" {
		secure = true
	}
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	// Validate the buckets name as being created by s3verify.
	bucketNameParts := strings.Split(bucketName, "-")
	if bucketNameParts[0] != "s3verify" {
		err := fmt.Errorf("%s is not an s3verify created bucket. See s3verify --help", bucketName)
		return err
	}
	validBucket := BucketInfo{
		Name: bucketName,
	}
	// Store the validated bucket in the global preparedBuckets array.
	preparedBuckets = append(preparedBuckets, validBucket)

	// Store the objects s3verify-object- inside this bucket inside the global object array.
	doneCh := make(chan struct{})
	objectInfoCh := client.ListObjects(bucketName, "s3verify/", true, doneCh)
	for objectInfo := range objectInfoCh {
		object := &ObjectInfo{
			// If objects are prepared they need to be given the correct metadata to pass listing tests.
			Key:          objectInfo.Key,
			ETag:         objectInfo.ETag,
			LastModified: objectInfo.LastModified,
		}
		preparedObjects = append(preparedObjects, object)
	}
	// Make sure that enough objects were actually found with the right prefix.
	if len(preparedObjects) < numTestObjects {
		err := fmt.Errorf("Not enough test objects found: need at least %d, only found %d", numTestObjects, len(preparedObjects))
		return err
	}
	return nil
}

// TODO: Create function using minio-go to upload 1001 parts of a multipart operation.

// mainPrepareS3Verify - Create one new buckets and 1001 objects for s3verify to use in the test.
func mainPrepareS3Verify(config ServerConfig) (string, error) {
	// Extract necessary values from the config.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return "", err
	}
	region := config.Region
	secure := false
	if hostURL.Scheme == "https" {
		secure = true
	}
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return "", err
	}
	// Create testing buckets.
	validBucketName, err := prepareBuckets(region, client)
	if err != nil {
		return "", err
	}
	// Use the first newly created bucket to store all the objects.
	if err := prepareObjects(client, validBucketName); err != nil {
		return "", err
	}
	return validBucketName, nil
}
