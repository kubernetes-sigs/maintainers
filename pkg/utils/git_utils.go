/*
Copyright 2022 The Kubernetes Authors.

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

package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckoutAtDate checks out the commit at the specified date.
func CheckoutAtDate(branch, date, dir string) error {
	err := os.Chdir(dir)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"/bin/bash",
		"-c",
		fmt.Sprintf("git checkout `git rev-list -n 1 --before=\"%s\" %s`", date, branch),
	)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func GetBranchName(dir string) (string, error) {
	err := os.Chdir(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(
		"/bin/bash",
		"-c",
		"git rev-parse --abbrev-ref HEAD",
	)

	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func Checkout(branch, dir string) error {
	err := os.Chdir(dir)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"/bin/bash",
		"-c",
		fmt.Sprintf("git checkout %s", branch),
	)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
