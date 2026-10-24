package main

import (
	"fmt"
	"os"
)

// interactiveSavePaths 只供互動入口使用；批次重播仍須明確指定輸入。
func interactiveSavePaths(writePath, readPath string, explicitLoad bool) (string, string, error) {
	if writePath == "" {
		if readPath != "" {
			writePath = readPath
		} else {
			var err error
			writePath, err = defaultSavePath()
			if err != nil {
				return "", "", err
			}
		}
	}
	if readPath != "" || explicitLoad {
		return writePath, readPath, nil
	}
	info, err := os.Stat(writePath)
	if os.IsNotExist(err) {
		return writePath, "", nil
	}
	if err != nil {
		return "", "", err
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("存檔不是一般檔案：%s", writePath)
	}
	return writePath, writePath, nil
}
