#
# Copyright 2022 SkyAPM org
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

TEST_SHELL="./tools/test.sh"
PLUGIN_DIR?=''

.PHONY: test
test:
	${TEST_SHELL} test ${PLUGIN_DIR}

.PHONY: deps
deps:
	${TEST_SHELL} deps  ${PLUGIN_DIR}

LINTER := bin/golangci-lint
$(LINTER):
	curl -L https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.31.0

.PHONY: lint
lint: $(LINTER)
	${TEST_SHELL} lint ${PLUGIN_DIR}

.PHONY: fix
fix: $(LINTER)
	${TEST_SHELL} fix ${PLUGIN_DIR}

