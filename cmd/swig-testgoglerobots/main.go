package main

import (
	"fmt"

	"github.com/msk-siteimprove/conn-checker/external/googlerobotstxt"
)

// #cgo LDFLAGS: -L/usr/lib/ -lstdc++ -l../../external/googlerobotstxt/robotstxt/c-build/librobots.a
// #include "../../robotstxt/c-build/libs/abseil-cpp-src/absl/strings/string_view.h"
// #cgo CXXFLAGS: -I/usr/lib/

// #cgo CXXFLAGS: -std=c++17
// #cgo LDFLAGS: -L. -l:../../external/googlerobotstxt/robotstxt/c-build/librobots.a
import "C"

func main() {
	fmt.Println("hello")
	matcher := googlerobotstxt.NewRobotsMatcher()
	a := matcher.Disallow()
	fmt.Println(a)
	// matcher.Disallow()
	// matcher.AllowedByRobots()
	// res := matcher.AllowedByRobots()

	// a := googlerobotstxt.SwigcptrAbsl_string_view(1)
	// b := googlerobotstxt.Absl_string_view(a)
	// googlerobotstxt.AllowedByRobots(b, googlerobotstxt.NewStringVector(), "")

	// fmt.Println(b)
}