// +build tools

package tools

import (
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.

// See:
// https://github.com/maxbrunsfeld/counterfeiter#step-1---create-toolsgo
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
