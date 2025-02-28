// @File     : file
// @Author   : Ville
// @Time     : 19-10-15 上午11:08
// files
package files

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// PathExists 检查文件路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// WriteToExecuteDir  写入数据到当前运行程序目录下某个位置
func WriteToExecuteDir(name string, content string) (dir string, err error) {
	dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return
	}
	dir = path.Join(dir, name)
	var file *os.File
	// 文件不存在使用默认配置
	if file, err = os.OpenFile(dir, os.O_RDWR|os.O_CREATE, 0666); err != nil {
		return
	}
	defer file.Close()
	buf := strings.NewReader(content)
	_, err = io.Copy(file, buf)
	return
}

// Remove 移除指定文件
func Remove(name string) error {
	if PathExists(name) {
		return os.Remove(name)
	}
	return nil
}
