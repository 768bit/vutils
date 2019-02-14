package vutils

import (
	"errors"
	"github.com/bmatcuk/doublestar"
	"os"
	"path/filepath"
	"strings"
)

type SourcePath = string
type DestPath = string

func NewContentsMap(resolveSymboilcLinks bool) *ContentsMap {
	return &ContentsMap{
		dirs:                 map[SourcePath]*ContentsMapDestinationDirectory{},
		files:                map[SourcePath]*ContentsMapDestinationFile{},
		resolveSymboilcLinks: resolveSymboilcLinks,
	}
}

type ContentsMap struct {
	dirs                 map[SourcePath]*ContentsMapDestinationDirectory
	files                map[SourcePath]*ContentsMapDestinationFile
	resolveSymboilcLinks bool
}

type ContentsMapDestinationFile struct {
	sourcePath        string
	mode              os.FileMode
	destPath          DestPath
	symbolicPath      string
	isSymbolicResolve bool
}

type ContentsMapDestinationDirectory struct {
	sourcePath        string
	mode              os.FileMode
	destPath          DestPath
	include           []string
	exclude           []string
	recursive         bool
	symbolicPathRoot  string
	isSymbolicResolve bool
}

func (cm *ContentsMap) AddDirectory(sourcePath string, destPath string, mode os.FileMode, recursive bool) error {
	p := filepath.Clean(string(sourcePath))
	if !Files.PathExists(p) {
		return os.ErrNotExist
	}
	cm.dirs[p] = &ContentsMapDestinationDirectory{
		sourcePath: p,
		mode:       mode,
		destPath:   destPath,
		recursive:  recursive,
	}
	return nil
}

func (cm *ContentsMap) AddDirectoryInclude(sourcePath string, destPath string, mode os.FileMode, recursive bool, include []string) error {
	p := filepath.Clean(string(sourcePath))
	if !Files.PathExists(p) {
		return os.ErrNotExist
	}
	cm.dirs[p] = &ContentsMapDestinationDirectory{
		sourcePath: p,
		mode:       mode,
		destPath:   destPath,
		recursive:  recursive,
		include:    include,
	}
	return nil
}

func (cm *ContentsMap) AddDirectoryExclude(sourcePath string, destPath string, mode os.FileMode, recursive bool, exclude []string) error {
	p := filepath.Clean(string(sourcePath))
	if !Files.PathExists(p) {
		return os.ErrNotExist
	}
	cm.dirs[sourcePath] = &ContentsMapDestinationDirectory{
		sourcePath: p,
		mode:       mode,
		destPath:   destPath,
		recursive:  recursive,
		exclude:    exclude,
	}
	return nil
}

func (cm *ContentsMap) AddFile(sourcePath string, destPath string, mode os.FileMode) error {
	p := filepath.Clean(string(sourcePath))
	if !Files.PathExists(p) {
		return os.ErrNotExist
	}
	cm.files[p] = &ContentsMapDestinationFile{
		sourcePath: p,
		mode:       mode,
		destPath:   destPath,
	}
	return nil
}

func (cm *ContentsMap) DoCopy(destRoot string) error {
	for _, dirItem := range cm.dirs {

		if err := cm.doDirCopy(destRoot, dirItem, []string{}); err != nil {
			return err
		}

	}

	for _, fileItem := range cm.files {

		if err := cm.doFileCopy(destRoot, fileItem); err != nil {
			return err
		}

	}

	return nil
}

func (cm *ContentsMap) doDirCopy(destRoot string, dirItem *ContentsMapDestinationDirectory, builtPath []string) error {

	return filepath.Walk(cm.makeDirSourcePath(dirItem.sourcePath, builtPath), func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			if !dirItem.recursive {
				return nil
			}

			if item, ok := cm.dirs[path]; ok && item != nil {

				return nil

			}
		} else {

			if item, ok := cm.files[path]; ok && item != nil {

				return nil

			}

		}

		//does the found item exist in the source content map?

		var excluded = false
		var ruleErr error

		if dirItem.exclude != nil && len(dirItem.exclude) > 0 {
			//do we need to exclude the directory?

			excluded, ruleErr = cm.runExcludeIncludeRules(destRoot, path, dirItem, builtPath, info, false)
			if ruleErr != nil {
				return ruleErr
			}

		}

		if !excluded && dirItem.include != nil && len(dirItem.include) > 0 {

			excluded, ruleErr = cm.runExcludeIncludeRules(destRoot, path, dirItem, builtPath, info, true)
			if ruleErr != nil {
				return ruleErr
			}

		}

		if !excluded {
			if info.IsDir() {
				if err := cm.doDirCopy(destRoot, dirItem, append(builtPath, info.Name())); err != nil {
					return err
				}
			} else if info.Mode()&os.ModeSymlink != 0 {
				superPath := path
				childDestRoot := destRoot
				if dirItem.isSymbolicResolve && dirItem.symbolicPathRoot != "" {
					if len(builtPath) == 0 {
						superPath = dirItem.symbolicPathRoot
					} else {
						superPath = filepath.Join(dirItem.symbolicPathRoot, filepath.Join(builtPath...))
						childDestRoot = filepath.Join(childDestRoot, filepath.Join(builtPath...))
					}
				}
				return cm.handleSymbolicLinkCopy(childDestRoot, dirItem, builtPath, path, info, superPath)
			} else {
				fmode := dirItem.mode
				if fmode == 0 {
					fmode = info.Mode()
				}
				if err := fcopyMode(path, cm.makeDestPath(destRoot, builtPath, info.Name()), info, fmode); err != nil {
					return err
				}
			}
		}

		return nil

	})

}

func (cm *ContentsMap) handleSymbolicLinkCopy(destRoot string, dirItem *ContentsMapDestinationDirectory, builtPath []string, path string, info os.FileInfo, superPath string) error {
	linkSrc, linkInfo, err := resolvelink(path)
	if err != nil {
		return err
	}

	if cm.resolveSymboilcLinks {

		if linkInfo.IsDir() {

			if !dirItem.recursive {
				return nil
			}

			if item, ok := cm.dirs[linkSrc]; ok && item != nil {

				return nil

			}

			lmode := dirItem.mode
			if lmode == 0 {
				if info.Mode() == 0 {
					if linkInfo.Mode() == 0 {
						lmode = os.ModePerm
					} else {
						lmode = linkInfo.Mode()
					}
				} else {
					lmode = info.Mode()
				}
			}
			if err := cm.doDirCopy(destRoot, &ContentsMapDestinationDirectory{
				sourcePath:        linkSrc,
				mode:              lmode,
				destPath:          cm.makeDestPath(destRoot, builtPath, info.Name()),
				symbolicPathRoot:  superPath,
				include:           dirItem.include,
				exclude:           dirItem.exclude,
				isSymbolicResolve: true,
			}, []string{}); err != nil {
				return err
			}

		} else if linkInfo.Mode()&os.ModeSymlink != 0 {

			return cm.handleSymbolicLinkCopy(destRoot, dirItem, builtPath, linkSrc, info, superPath)

		} else {
			//a file
			fmode := dirItem.mode
			if fmode == 0 {
				if info.Mode() == 0 {
					if linkInfo.Mode() == 0 {
						fmode = os.ModePerm
					} else {
						fmode = linkInfo.Mode()
					}
				} else {
					fmode = info.Mode()
				}
			}
			if err := fcopyMode(path, cm.makeDestPath(destRoot, builtPath, info.Name()), info, fmode); err != nil {
				return err
			}

		}

	} else {

		//standard symbolic link...

		lmode := dirItem.mode
		if lmode == 0 {
			if info.Mode() == 0 {
				if linkInfo.Mode() == 0 {
					lmode = os.ModePerm
				} else {
					lmode = linkInfo.Mode()
				}
			} else {
				lmode = info.Mode()
			}
		}

		if err := lcopyMode(path, cm.makeDestPath(destRoot, builtPath, info.Name()), info, lmode); err != nil {
			return err
		}

	}
	return nil
}

func (cm *ContentsMap) handleSymbolicLinkFileCopy(destRoot string, fileItem *ContentsMapDestinationFile, path string, info os.FileInfo, superPath string) error {
	linkSrc, linkInfo, err := resolvelink(path)
	if err != nil {
		return err
	}

	if cm.resolveSymboilcLinks {

		if linkInfo.IsDir() {

			return errors.New("Cannot copy a directory using symbolic link file copy methods")

		} else if linkInfo.Mode()&os.ModeSymlink != 0 {

			return cm.handleSymbolicLinkFileCopy(destRoot, fileItem, linkSrc, info, superPath)

		} else {
			//
			fmode := fileItem.mode
			if fmode == 0 {
				if info.Mode() == 0 {
					if linkInfo.Mode() == 0 {
						fmode = os.ModePerm
					} else {
						fmode = linkInfo.Mode()
					}
				} else {
					fmode = info.Mode()
				}
			}
			if err := fcopyMode(path, cm.makeDestFilePath(destRoot, fileItem.destPath), info, fmode); err != nil {
				return err
			}
		}

	} else {

		//standard symbolic link...

		lmode := fileItem.mode
		if lmode == 0 {
			if info.Mode() == 0 {
				if linkInfo.Mode() == 0 {
					lmode = os.ModePerm
				} else {
					lmode = linkInfo.Mode()
				}
			} else {
				lmode = info.Mode()
			}
		}

		if err := lcopyMode(path, cm.makeDestFilePath(destRoot, fileItem.destPath), info, lmode); err != nil {
			return err
		}

	}
	return nil
}

func (cm *ContentsMap) runExcludeIncludeRules(destRoot string, path string, dirItem *ContentsMapDestinationDirectory, builtPath []string, info os.FileInfo, isIncludeMode bool) (bool, error) {
	excluded := true

	var rules []string
	if isIncludeMode {
		rules = dirItem.include
	} else {
		rules = dirItem.exclude
	}

	matchSet := []string{
		info.Name(),
	}

	actualSourcePath := dirItem.sourcePath

	if dirItem.isSymbolicResolve {
		matchSet = append(matchSet, cm.makeSourcePath(dirItem.symbolicPathRoot, builtPath, info.Name()))
		actualSourcePath = dirItem.symbolicPathRoot
	} else {
		matchSet = append(matchSet, path)
	}

	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		cleanRule := rule
		fullRule := cleanRule
		if cleanRule[0] == '/' {
			if len(cleanRule) == 1 {
				continue
			}
			cleanRule = strings.Replace(cleanRule, actualSourcePath, "", 1)
			if len(cleanRule) == 0 || cleanRule == "/" || cleanRule == "./" {
				if isIncludeMode {
					return false, nil
				}
				return true, nil
			}
			if cleanRule[0] == '/' {
				cleanRule = cleanRule[1:]
			} else if cleanRule[0:1] == "./" {
				cleanRule = cleanRule[2:]
			}
			fullRule = filepath.Join(actualSourcePath, cleanRule)
		} else if cleanRule[0:1] == "./" {
			if len(cleanRule) == 2 || cleanRule == "./" {
				if isIncludeMode {
					return false, nil
				}
				return true, nil
			}
			cleanRule = cleanRule[2:]
			fullRule = filepath.Join(actualSourcePath, cleanRule)
		} else {
			cleanRule = strings.Replace(cleanRule, actualSourcePath, "", 1)
			if len(cleanRule) == 0 || cleanRule == "/" || cleanRule == "./" {
				if isIncludeMode {
					return false, nil
				}
				return true, nil
			}
			if cleanRule[0] == '/' {
				cleanRule = cleanRule[1:]
			} else if cleanRule[0:1] == "./" {
				cleanRule = cleanRule[2:]
			}
			fullRule = filepath.Join(actualSourcePath, cleanRule)
		}

		if !dirItem.recursive && (strings.Contains(cleanRule, "/") || strings.Contains(cleanRule, "**")) {
			continue
		} else if !info.IsDir() {

			for _, matchItem := range matchSet {

				if matches, err := doublestar.Match(fullRule, matchItem); err != nil {
					return true, err
				} else if !matches {
					if isIncludeMode {
						excluded = true
						continue
					}
					excluded = false
				} else {
					if isIncludeMode {
						excluded = false
						return excluded, nil
					}
					excluded = true
					return excluded, nil
				}

			}

			if !isIncludeMode && excluded {
				return excluded, nil
			}

		} else if dirItem.recursive {

			for _, matchItem := range matchSet {

				if matches, err := doublestar.Match(fullRule, matchItem); err != nil {
					return true, err
				} else if !matches {
					if isIncludeMode {
						excluded = true
						continue
					}
					excluded = false
				} else {
					if isIncludeMode {
						excluded = false
						return excluded, nil
					}
					excluded = true
					return excluded, nil
				}

			}

			if !isIncludeMode && excluded {
				return excluded, nil
			}

		} else {
			return true, nil
		}

	}

	return excluded, nil

}

func (cm *ContentsMap) doFileCopy(destRoot string, fileItem *ContentsMapDestinationFile) error {
	info, err := os.Lstat(fileItem.sourcePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("Cannot copy a directory using file copy methods")
	} else if info.Mode()&os.ModeSymlink != 0 {
		superPath := fileItem.sourcePath
		childDestRoot := destRoot
		return cm.handleSymbolicLinkFileCopy(childDestRoot, fileItem, fileItem.sourcePath, info, superPath)
	} else {
		fmode := fileItem.mode
		if fmode == 0 {
			fmode = info.Mode()
		}
		if err := fcopyMode(fileItem.sourcePath, cm.makeDestFilePath(destRoot, fileItem.destPath), info, fmode); err != nil {
			return err
		}
	}

	return nil

}

func (cm *ContentsMap) makeSourcePath(sourcePath string, builtPath []string, name string) string {
	resSourcePath := sourcePath
	if len(builtPath) > 0 {
		resSourcePath = filepath.Join(resSourcePath, filepath.Join(builtPath...))
	}
	return filepath.Join(resSourcePath, name)
}

func (cm *ContentsMap) makeDirSourcePath(sourcePath string, builtPath []string) string {
	resSourcePath := sourcePath
	if len(builtPath) > 0 {
		resSourcePath = filepath.Join(resSourcePath, filepath.Join(builtPath...))
	}
	return resSourcePath
}

func (cm *ContentsMap) makeDestPath(destRoot string, builtPath []string, dest string) string {
	destPath := destRoot
	if len(builtPath) > 0 {
		destPath = filepath.Join(destPath, filepath.Join(builtPath...))
	}
	return filepath.Join(destPath, dest)
}

func (cm *ContentsMap) makeDestFilePath(destRoot string, dest string) string {
	destPath := destRoot
	if dest[0] == '/' {
		destPath = filepath.Join(destPath, destPath[1:])
	}
	return filepath.Join(destPath, dest)
}
