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

// prepareBuckets - Uses minio-go library to create new testing buckets for use by s3verify.
func prepareBuckets(region string, client *minio.Client) ([]string, error) {
	message := "Creating test buckets"
	bucketNames := []string{
		randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-"),
		randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-"),
	}
	for _, bucketName := range bucketNames {
		// Spin scanBar
		scanBar(message)
		err := client.MakeBucket(bucketName, region)
		if err != nil {
			printMessage(message, err)
			return nil, err
		}
		// Spin scanBar
		scanBar(message)
	}
	// Bucket preparation passed.
	printMessage(message, nil)
	return bucketNames, nil
}

// TODO: see if parallelization has a place here.

// prepareObjects - Uses minio-go library to create 1001 new testing objects for use by s3verify.
func prepareObjects(client *minio.Client, bucketName string) error {
	message := "Creating test objects"
	// TODO: update this to 1001...for testing purposes it is OK to leave it at 101 for now.
	// Upload 1001 objects specifically for the list-objects tests.
	for i := 0; i < 101; i++ {
		// Spin scanBar
		scanBar(message)
		randomData := randString(60, rand.NewSource(time.Now().UnixNano()), "")
		objectKey := "s3verify-object-" + strconv.Itoa(i)
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
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, true)
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
	// Store the validated bucket in the global unpreparedBuckets array.
	unpreparedBuckets = append(unpreparedBuckets, validBucket)

	// Store the objects s3verify-object- inside this bucket inside the global object array.
	doneCh := make(chan struct{})
	objectInfoCh := client.ListObjects(bucketName, "s3verify-object-", true, doneCh)
	for objectInfo := range objectInfoCh {
		object := &ObjectInfo{
			Key: objectInfo.Key,
		}
		objects = append(objects, object)
	}
	return nil
}

// TODO: Create function using minio-go to upload 1001 parts of a multipart operation.

// mainPrepareS3Verify - Create two new buckets and 1001 objects for s3verify to use in the test.
func mainPrepareS3Verify(config ServerConfig) ([]string, error) {
	// Extract necessary values from the config.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	region := config.Region
	// Create a new Minio-Go client object.
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, true)
	if err != nil {
		return nil, err
	}
	// Create testing buckets.
	validBucketNames, err := prepareBuckets(region, client)
	if err != nil {
		return nil, err
	}
	// Use the first newly created bucket to store all the objects.
	if err := prepareObjects(client, validBucketNames[0]); err != nil {
		return nil, err
	}
	return validBucketNames, nil
}
