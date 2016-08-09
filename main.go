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
	"fmt"
	"os"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

// Global scanBar for all tests to access and update.
var scanBar = scanBarFactory()

// Global command line flags.
var (
	s3verifyFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Show help.",
		},
	}
)

// Custom help template.
// Revert to API not command.
var s3verifyHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}

USAGE:
	{{.Name}} {{if .Flags}}[FLAGS...] {{end}}
GLOBAL FLAGS:
	{{range .Flags}}{{.}}
	{{end}}
EXAMPLES:
	1. Run all tests on Minio server. Note play.minio.io:9000 is a public test server. 
	You are free to use these Secret and Access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify --extended
	2. Run all basic tests on Amazon S3 server using flags. 
	Note that passing access and secret keys as flags should be avoided on a multi-user server for security reasons.
		$ s3verify --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com --region us-west-1
`

// APItest - Define all mainXXX tests to be of this form.
type APItest struct {
	Test     func(ServerConfig, int) bool
	Extended bool // Extended tests will only be invoked at the users request.
	Critical bool // Tests marked critical must pass before more tests can be run.
}

func commandNotFound(ctx *cli.Context, command string) {
	msg := fmt.Sprintf("'%s' is not a s3verify command. See 's3verify --help'.", command)
	console.PrintC(msg)
}

// registerApp - Create a new s3verify app.
func registerApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "Test for Amazon S3 v4 API compatibility."
	app.Author = "Minio.io"
	app.Name = "s3verify"
	app.Flags = append(s3verifyFlags, globalFlags...)
	app.CustomAppHelpTemplate = s3verifyHelpTemplate // Custom help template defined above.
	app.CommandNotFound = commandNotFound            // Custom commandNotFound function defined above.
	app.Action = callAllAPIs                         // Command to run if no commands are explicitly passed.
	return app
}

// makeConfigFromCtx - parse the passed context to create a new config.
func makeConfigFromCtx(ctx *cli.Context) (*ServerConfig, error) {
	if ctx.GlobalString("access") != "" &&
		ctx.GlobalString("secret") != "" &&
		ctx.GlobalString("url") != "" {
		config := newServerConfig(ctx)
		return config, nil
	}
	// If config cannot be created successfully show help and exit immediately.
	return nil, fmt.Errorf("Unable to create config.")
}

// callAllAPIS parse context extract flags and then call all.
func callAllAPIs(ctx *cli.Context) {
	// Create a new config from the context.
	config, err := makeConfigFromCtx(ctx)
	if err != nil {
		// Could not create a config. Exit immediately.
		cli.ShowAppHelpAndExit(ctx, 1)
	}
	// Test that the given endpoint is reachable with a simple GET request.
	if err := verifyHostReachable(config.Endpoint, config.Region); err != nil { // If the provided endpoint is unreachable error out instantly.
		console.Fatalln(err)
	}
	if ctx.GlobalBool("prepare") {
		// Create a prepared testing environment with 2 buckets and 1001 objects and 1001 object parts.
		bucketNames, err := mainPrepareS3Verify(*config)
		if err != nil {
			console.Fatalln(err)
		}
		console.Printf("Please run 's3verify -a [YOUR_ACCESS_KEY] -s [YOUR_SECRET_KEY] -u [HOST_URL] [FLAGS...] %s %s'\n", bucketNames[0], bucketNames[1])
	} else if ctx.GlobalString("clean") != "" {
		if err := cleanObjects(*config, ctx.GlobalString("clean")); err != nil {
			console.Fatalln(err)
		}
		if err := cleanBucket(*config, ctx.GlobalString("clean")); err != nil {
			console.Fatalln(err)
		}
	} else if len(ctx.Args()) == 2 {
		console.Printf("s3verify attempting to test AWS signv4 compatability using %s and %s...\n", ctx.Args()[0], ctx.Args()[1])
		// Validate the passed names as valid bucket names and store their objects in the global objects array.
		for _, bucketName := range ctx.Args() {
			validateBucket(*config, bucketName)
		}
		testCount := 1
		for _, test := range preparedTests { // Run all tests that have been set up.
			if test.Extended {
				if ctx.GlobalBool("extended") {
					test.Test(*config, testCount)
					testCount++
				}
			} else {
				if !test.Test(*config, testCount) && test.Critical {
					os.Exit(1)
				}
				testCount++
			}
		}
	} else {
		// If the user does not use --prepare flag then just run all non preparedTests.
		testCount := 1
		for _, test := range unpreparedTests {
			if test.Extended {
				if ctx.GlobalBool("extended") {
					test.Test(*config, testCount)
				}
			} else {
				if !test.Test(*config, testCount) && test.Critical {
					os.Exit(1)
				}
				testCount++
			}
		}
	}
}

// main - Set up and run the app.
func main() {
	app := registerApp()
	app.Before = func(ctx *cli.Context) error {
		setGlobalsFromContext(ctx)
		return nil
	}
	app.RunAndExitOnError()
}
