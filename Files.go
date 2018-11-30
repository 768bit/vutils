package vutils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func (fu *filesUtils) PathExists(path string) bool {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}
	return true

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

func (fu *filesUtils) AppendStringToFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

// Copy copies src to dest, doesn't matter if src is a directory or a file
func (fu *filesUtils) Copy(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	return rcopy(src, dest, info)
}

// Copy copies src to dest, doesn't matter if src is a directory or a file, deletes src after completion
func (fu *filesUtils) CopyRM(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	err = rcopy(src, dest, info)
	if err == nil {
		return os.RemoveAll(src)
	}
	return err
}

// copy dispatches copy-funcs according to the mode.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func rcopy(src, dest string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return lcopy(src, dest, info)
	}
	if info.IsDir() {
		return dcopy(src, dest, info)
	}
	return fcopy(src, dest, info)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

// dcopy is for a directory,
// with scanning contents inside the directory
// and pass everything to "copy" recursively.
func dcopy(srcdir, destdir string, info os.FileInfo) error {

	if err := os.MkdirAll(destdir, info.Mode()); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := rcopy(cs, cd, content); err != nil {
			// If any error, exit immediately
			return err
		}
	}
	return nil
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string, info os.FileInfo) error {
	src, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

var Files = &filesUtils{}
