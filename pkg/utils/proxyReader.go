package utils

import (
	"os"
	"strings"
)

func ReadProxyList(proxyFilePath string) ([]string, error) {
	proxyList, err := os.ReadFile(proxyFilePath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(proxyList), "\n"), nil
}
