// Copyright Istio Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package branch

import (
	"bufio"
	"bytes"
	"fmt"

	"istio.io/pkg/log"
	"istio.io/release-builder/pkg/model"
	"istio.io/release-builder/pkg/util"
)

// UpdateCommonFilesCommon update the common-files repo for the new release.
// It will search for the latest build-tools image, and place it in IMAGE_VERSION
// as well as update the UPDATE_BRANCH.
// A prereq for this is that the common-files release branch has been updated with a
// new UPDATE_BRANCH and image in it's files.
func UpdateCommonFilesCommon(manifest model.Manifest, release string, dryrun bool) error {
	log.Infof("*** Updating common-files")
	repo := "common-files"

	log.Infof("***Updating the common-files for %s from directory: %s", repo, manifest.RepoDir(repo))
	sedString := "s/UPDATE_BRANCH ?=.*/UPDATE_BRANCH ?= \"release-" + release + "\"/"
	cmd := util.VerboseCommand("sed", "-i", sedString, "files/common/Makefile.common.mk")
	cmd.Dir = manifest.RepoDir(repo)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	// In this command string we get the list of tags for the build-tools image,
	// awk those containing release-<release> and not latest, and then sort in reverse
	// order to get newest at the top. Tag is the first line.
	cmdString := "curl -sL https://gcr.io/v2/istio-testing/build-tools/tags/list | jq '.\"manifest\"[][\"tag\"]' | " +
		" awk '/release-" + release + "/ && !/latest/' | sort -r | sed  -e s/[[:space:]]*\\\"// -e s/\\\".*//"
	cmd = util.VerboseCommand("bash", "-c", cmdString)
	cmd.Stdout = nil
	cmd.Dir = manifest.RepoDir(repo)
	var tagBytes []byte
	var err error
	if tagBytes, err = cmd.Output(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}
	tag, _, _ := (bufio.NewReader(bytes.NewReader(tagBytes))).ReadLine()

	sedString = "s/IMAGE_VERSION=.*/IMAGE_VERSION=" + string(tag) + "/"
	cmd = util.VerboseCommand("sed", "-i", sedString, "files/common/scripts/setup_env.sh")
	cmd.Dir = manifest.RepoDir(repo)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	log.Infof("*** common-files updated")
	return nil
}
