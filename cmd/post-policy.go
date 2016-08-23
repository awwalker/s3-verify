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

import "time"

const expirationDateFormat = "2006-01-02T15:04:05.999Z"

// TODO: so far this function only creates valid policies.
// Should create some invalid cases to check for error messages.

// newPostPolicyBytes - creates a bare bones postpolicy string with key and bucket matches.
func newPostPolicyBytes(credential, bucketName, objectKey string, expiration time.Time) []byte {
	t := time.Now().UTC()
	expirationStr := `"expiration":` + `"` + expiration.Format(expirationDateFormat) + `"`
	bucketConditionStr := `["eq", "$bucket", ` + `"` + bucketName + `"]`
	keyConditionStr := `["eq", "$key", "` + objectKey + `"]`
	algorithmConditionStr := `["eq", "$x-amz-algorithm", "AWS4-HMAC-SHA256"]`
	dateConditionStr := `["eq", "$x-amz-date",` + `"` + t.Format(iso8601DateFormat) + `"]`
	credentialConditionStr := `["eq", "$x-amz-credential",` + `"` + credential + `"]`

	conditionStr := `"conditions":[` + bucketConditionStr + "," + keyConditionStr + "," + algorithmConditionStr + "," + dateConditionStr + "," + credentialConditionStr + "]"
	retStr := "{"
	retStr = retStr + expirationStr + ","
	retStr = retStr + conditionStr
	retStr = retStr + "}"

	return []byte(retStr)
}
