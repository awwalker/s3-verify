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

import "time"

// BucketInfo container for bucket metadata.
type BucketInfo struct {
	// The name of the bucket.
	Name string `json:"name"`
	// Date the bucket was created.
	CreationDate time.Time `json:"creationDate"`
}

// ObjectInfo container for object metadata.
type ObjectInfo struct {
	// An ETag is optionally set to md5sum of an object.  In case of multipart objects,
	// ETag is of the form MD5SUM-N where MD5SUM is md5sum of all individual md5sums of
	// each parts concatenated into one string.
	ETag string `json:"etag"`

	Key          string    `json:"name"`         // Name of the object
	LastModified time.Time `json:"lastModified"` // Date and time the object was last modified.
	Size         int64     `json:"size"`         // Size in bytes of the object.
	ContentType  string    `json:"contentType"`  // A standard MIME type describing the format of the object data.

	// Owner name.
	Owner struct {
		DisplayName string `json:"name"`
		ID          string `json:"id"`
	} `json:"owner"`

	// The class of storage used to store the object.
	StorageClass string `json:"storageClass"`

	// Error
	Err error `json:"-"`

	Body     []byte // Data held by the object.
	UploadID string // To be set only for multipart uploaded objects.
}

// A container for ObjectInfo structs to allow sorting.
type ObjectInfos []ObjectInfo

// Return the len of a list of ObjectInfo.
func (o ObjectInfos) Len() int {
	return len(o)
}

// Allow comparisons of ObjectInfo types with their Keys.
func (o ObjectInfos) Less(i, j int) bool {
	return o[i].Key < o[j].Key
}

// Allow swapping of ObjectInfo types.
func (o ObjectInfos) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

// objectInfoChannel a channel for concurrent object level operations.
type objectInfoChannel struct {
	objInfo ObjectInfo
	err     error
	index   int
}

// partChannel a channel for concurrent multipart upload part operations.
type partChannel struct {
	objPart objectPart
	err     error
	index   int
}

// multiUploadInitChannel a channel for concurrent multipart initiate upload operations.
type multiUploadInitChannel struct {
	uploadID string
	err      error
	index    int
}
