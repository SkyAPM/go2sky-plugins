//
// Copyright 2022 SkyAPM org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	go2sky = "github.com/SkyAPM/go2sky@v1.4.0"
)

func main() {
	pwd, _ := os.Getwd()
	err := scanGoMod(pwd)
	if err != nil {
		panic(err)
	}
}

func editGoMod(pluginPath string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && go mod edit -require=%s && go mod tidy", pluginPath, go2sky))
	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}

func scanGoMod(basePath string) error {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	hasGoMod := false
	for _, file := range files {
		if file.Name() == "go.mod" {
			hasGoMod = true
			break
		}
	}

	if hasGoMod {
		fmt.Printf("🐶 edit %s go.mod ...\n", basePath)
		return editGoMod(basePath)
	}

	for _, file := range files {
		if file.IsDir() {
			err1 := scanGoMod(fmt.Sprintf("%s/%s", basePath, file.Name()))
			if err1 != nil {
				return err1
			}
		}
	}

	return nil
}
