# S3Verify
S3 Verify is a project aimed at verifying the S3 compatibility of AWS V4 signature based object storage systems.

## INSTALLATION
### Prerequisites
- A working Golang Environment. If you do not have a working Golang environment, please follow [Install Golang](https://github.com/minio/minio/blob/master/INSTALLGO.md).
- [Minio Server] (https://github.com/minio/minio/blob/master/README.md) (optional--only necessary if you wish to test your own Minio Server for S3 Compatibility.)

### From Source
Currently s3verify is only available to be downloaded from source. 

```$ go get -u github.com/minio/s3verify```

## APIs
Currently S3 Verify is under heavy development and is subject to breaking changes. Currently we support five different API checks:
* PUT Bucket (putbucket)
* PUT Object (putobject)
* MULTIPART Object
* HEAD Object (headobject)
* COPY Object (copyobject)
* GET Object (getobject)
* GET Service (listbuckets)
* DELETE Object (removeobject)
* DELETE Bucket (removebucket)

We look forward to adding support for all standard AWS APIs for both Bucket and Object based operations.

## CLI USAGE
When s3verify is supplied with acceptable flags or environment variables it will run all API tests one after another. See Examples for detailed instructions.

```
$ s3verify [FLAGS]
```

### Flags
``s3verify`` implements the following flags
```
    --help      -h      Provides documentation for a given command.
    --access    -a      Allows user to input their AWS access key.
    --secretkey -s      Allows user to input their AWS secret access key.
    --url       -u      Allows user to input the host URL of the server they wish to test.
    --region    -r      Allows user to change the region of the AWS host they are using. Please do not use 'us-east-1' with
                        AWS servers or automatic cleanup of test buckets and objects will fail. Defaults to 'us-east-1'.
    --verbose     -v      [Under development] Currently allows user to trace the HTTP requests and responses sent by s3verify.
    --extended          Allows user to decide whether to test only basic S3 compliance or to test full API compliance.
```

### Environment Variables
``s3verify`` also supports the following environment variables as a replacement for flags. In fact it is recommended that on multiuser systems that env. 
variables be used for security reasons.  
The following env. variables can be used to replace their corresponding flags.
```
    S3_ACCESS can be set to YOUR_ACCESS_KEY and replaces --access -a.
    S3_SECRET can be set to YOUR_SECRET_KEY and replaces --secret -s.
    S3_REGION can be set to the region of the AWS host and replaces --region -r.
    S3_URL can be set to the host URL of the server users wish to test and replaces --url -u.
```
## EXAMPLES
Use s3verify to check the AWS S3 V4 compatibility of the Minio test server (https://play.minio.io:9000) 
```
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io:9000 
```

Use s3verify to check the AWS S3 V4 compatibility of the Minio test server with all APIs.
```
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECERT_KEY https://play.minio.io:9000 --extended
```

If a test fails you can use the verbose flag (--verbose) to check the request and response formed by the test to see where it failed.
```
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io:9000 --verbose
```
