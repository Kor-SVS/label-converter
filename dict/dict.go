package dict

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

var (
	isValidPhoneDict = false
	phoneDict        *PhoneDict
)

func init() {
	var err error

	printErr := color.New(color.FgRed).PrintfFunc()

	curPath := getExecutablePath()

	phoneDictPath := filepath.Join(curPath, "dict.txt")
	phoneDictPath, err = filepath.Abs(phoneDictPath)
	if err == nil {
		var bytesPhoneDict []byte
		bytesPhoneDict, err = os.ReadFile(phoneDictPath)
		if err == nil {
			var pd *PhoneDict
			pd, err = NewPhoneDict(bytesPhoneDict)
			if err == nil {
				isValidPhoneDict = true
				phoneDict = pd
				fmt.Printf("[PhoneDict Load] Rule Count: %v\n", phoneDict.Count())
			}
		}
	}

	if err != nil {
		printErr("[PhoneDict Load Error] %v\n", err)
	}
}

func IsValidPhoneDict() bool {
	return isValidPhoneDict
}

func CurrentPhoneDict() *PhoneDict {
	return phoneDict
}
