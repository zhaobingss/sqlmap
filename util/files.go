package util

import (
	"io/ioutil"
	"os"
)

/// 递归获取某个文件夹下面的所有的文件
func GetAllFiles(dirPth string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			fs, err := GetAllFiles(dirPth + PthSep + fi.Name())
			if err != nil {
				return nil, err
			} else {
				files = append(files, fs...)
			}
		} else {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}

	return files, nil
}
