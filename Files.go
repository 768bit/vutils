package vutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type filesUtils struct {
}

func (fu *filesUtils) CheckPathExists(path string) bool {

	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false

}

func (fu *filesUtils) GetRelativeFolderPathWithDefault(rootPath string, findPath string, defaultPath string) (error, string) {

	//need to check if the findPath is relative, if it is we check it exists relative to rootPath
	// if it doesnt exist there we need to use defaultPath relative to rootPath and check it is there... if not return error
	cleanRootPath := filepath.Clean(rootPath)
	cleanDefaultPath := filepath.Clean(defaultPath)
	rootPathExists := fu.CheckPathExists(cleanRootPath)

	cleanFindPath := filepath.Clean(findPath)

	isAbsPath := filepath.IsAbs(findPath)
	defaultIsAbs := filepath.IsAbs(cleanDefaultPath)

	if isAbsPath {
		//path is absolute check it exists...
		if !fu.CheckPathExists(cleanFindPath) {
			//path doesnt exist lets try using the default path
			if defaultIsAbs && fu.CheckPathExists(cleanDefaultPath) {

				return nil, cleanDefaultPath

			} else if fullPath := filepath.Join(cleanRootPath, cleanDefaultPath); defaultPath != "" && cleanDefaultPath != "." &&
				!defaultIsAbs && rootPathExists && fu.CheckPathExists(fullPath) {

				return nil, fullPath

			} else {

				return errors.New(fmt.Sprintf("Unable to find absolute path %s and unable to find the default path %s either absolutely or relatively to %s", findPath, defaultPath, rootPath)), ""

			}
		} else {

			return nil, cleanFindPath

		}
	} else {
		//not an absolute path so lets find relative to root dir...
		fullFindPath := filepath.Join(cleanRootPath, cleanFindPath)
		if !fu.CheckPathExists(fullFindPath) {
			//path doesnt exist lets get the default path
			if defaultIsAbs && fu.CheckPathExists(cleanDefaultPath) {

				return nil, cleanDefaultPath

			} else if fullPath := filepath.Join(cleanRootPath, cleanDefaultPath); defaultPath != "" && cleanDefaultPath != "." &&
				!defaultIsAbs && rootPathExists && fu.CheckPathExists(fullPath) {

				return nil, fullPath

			} else {

				return errors.New(fmt.Sprintf("Unable to find path %s or default path %s either absolutely or relatively to %s", findPath, defaultPath, rootPath)), ""

			}
		} else {

			return nil, fullFindPath

		}
	}

}

func (fu *filesUtils) RemoveTempFolderContents(dir string) error {

	d, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	files, err := d.Readdir(-1)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (fu *filesUtils) CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

var Files = &filesUtils{}
