/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"debug/macho"
	"fmt"

	"github.com/gravitational/kingpin"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
)

type Args struct {
	logLevel      *string
	logJSON       *bool
	AppleUsername *string
	ApplePassword *string
	BinaryPaths   *[]string
}

func NewArgs() *Args {
	logLevelStrings := make([]string, 0, len(logrus.AllLevels))
	for _, level := range logrus.AllLevels {
		logLevelStrings = append(logLevelStrings, level.String())
	}

	args := &Args{
		logLevel:      kingpin.Flag("log-level", "Output logging level").Default(logrus.InfoLevel.String()).Enum(logLevelStrings...),
		logJSON:       kingpin.Flag("log-json", "Enable JSON logging").Default(fmt.Sprintf("%v", false)).Bool(),
		AppleUsername: kingpin.Flag("apple-username", "Apple Connect username used for notarization").Required().Envar("APPLE_USERNAME").String(),
		ApplePassword: kingpin.Flag("apple-password", "Apple Connect password used for notarization").Required().Envar("APPLE_PASSWORD").String(),
		BinaryPaths:   kingpin.Arg("binaries", "Path to Apple binaries for signing and notarization").Required().Action(binaryArgValidatiorAction).ExistingFiles(),
	}

	return args
}

func binaryArgValidatiorAction(pc *kingpin.ParseContext) error {
	for _, element := range pc.Elements {
		binaryPath := *element.Value
		err := verifyFileIsAppleBinary(binaryPath)
		if err != nil {
			return trace.Wrap(err, "failed to verify that %q is a valid Apple binary for signing", binaryPath)
		}
	}
	return nil
}

// Returns an error if the file is not a valid Apple binary
// Effectively does `file $BINARY | grep -ic 'mach-o'`
func verifyFileIsAppleBinary(filePath string) error {
	// First check to see if the binary is a typical mach-o binary.
	// If it's not, it could still be a multiarch "fat" mach-o binary,
	// so we try that next. If both fail then the file is not an Apple
	// binary.
	fileHandle, err := macho.Open(filePath)
	if err != nil {
		fatFileHandle, err := macho.OpenFat(filePath)
		if err != nil {
			return trace.Wrap(err, "the provided file %q is neither a normal or multiarch mach-o binary.", filePath)
		}

		err = fatFileHandle.Close()
		if err != nil {
			return trace.Wrap(err, "identified %q as a multiarch mach-o binary but failed to close the file handle", filePath)
		}
	} else {
		err := fileHandle.Close()
		if err != nil {
			return trace.Wrap(err, "identified %q as a mach-o binary but failed to close the file handle", filePath)
		}
	}

	return nil
}

func parseArgs() (*Args, error) {
	args := NewArgs()
	kingpin.Parse()

	// This needs to be called as soon as possible so that the logger can
	// be used when checking args
	parsedLogLevel, err := logrus.ParseLevel(*args.logLevel)
	if err != nil {
		// This should never be hit if kingpin is configured correctly
		return nil, trace.Wrap(err, "failed to parse logrus log level")
	}
	NewLoggerConfig(parsedLogLevel, *args.logJSON).setupLogger()

	logrus.Debugf("Successfully parsed args: %v", args)
	return args, nil
}
