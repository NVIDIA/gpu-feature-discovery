/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package lm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Labels defines a type for labels
type Labels map[string]string

// Labels also implements the Labeler interface
func (labels Labels) Labels() (Labels, error) {
	return labels, nil
}

// WriteToFile writes labels to the specified path. The file is written atomocally
func (labels Labels) WriteToFile(path string) error {
	if path == "" {
		_, err := labels.WriteTo(os.Stdout)
		return err
	}

	output := new(bytes.Buffer)
	if _, err := labels.WriteTo(output); err != nil {
		return fmt.Errorf("error writing labels to buffer: %v", err)
	}
	err := writeFileAtomically(path, output.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error atomically writing file '%s': %v", path, err)
	}
	return nil
}

// WriteTo writes labels to the specified writer
func (labels Labels) WriteTo(output io.Writer) (int64, error) {
	var total int64
	for k, v := range labels {
		n, err := fmt.Fprintf(output, "%s=%s\n", k, v)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func writeFileAtomically(path string, contents []byte, perm os.FileMode) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to retrieve absolute path of output file: %v", err)
	}

	absDir := filepath.Dir(absPath)
	tmpDir := filepath.Join(absDir, "gfd-tmp")

	err = os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(tmpDir)
		}
	}()

	tmpFile, err := ioutil.TempFile(tmpDir, "gfd-")
	if err != nil {
		return fmt.Errorf("fail to create temporary output file: %v", err)
	}
	defer func() {
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
		}
	}()

	err = ioutil.WriteFile(tmpFile.Name(), contents, perm)
	if err != nil {
		return fmt.Errorf("error writing temporary file '%v': %v", tmpFile.Name(), err)
	}

	err = os.Rename(tmpFile.Name(), path)
	if err != nil {
		return fmt.Errorf("error moving temporary file to '%v': %v", path, err)
	}

	err = os.Chmod(path, perm)
	if err != nil {
		return fmt.Errorf("error setting permissions on '%v': %v", path, err)
	}

	return nil
}
