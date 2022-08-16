package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func DirExists(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.IsDir()
}

func Parse(filePath string, v interface{}) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := json.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
	default:
		return errors.New("unsupported file type")
	}

	return nil
}
