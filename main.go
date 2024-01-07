package main

import (
	"errors"
	"fmt"
	"labelconverter/dict"
	"labelconverter/lab"
	"labelconverter/label"
	"labelconverter/textgrid"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/sqweek/dialog"
)

const (
	_100_NANO_SEC = 10000000
)

func checkErr(err error) {
	if err != nil {
		printErr := color.New(color.FgRed).PrintfFunc()
		printErr("[ERROR] %v\n", err)
		fmt.Scanln()
		panic(err)
	}
}

func checkIsFilePath(fpath string) bool {
	stat, err := os.Stat(fpath)
	return !errors.Is(err, os.ErrNotExist) && !stat.IsDir()
}

func checkIsDirPath(fpath string) bool {
	stat, err := os.Stat(fpath)
	return !errors.Is(err, os.ErrNotExist) && stat.IsDir()
}

func openDialog() string {
	dialogBuilder := dialog.Directory()
	dialogBuilder = dialogBuilder.Title("Select Label Dir Path")

	dirPath, err := os.Getwd()
	if err == nil {
		dialogBuilder = dialogBuilder.SetStartDir(dirPath)
	}

	directory, err := dialogBuilder.Browse()
	if err == dialog.Cancelled {
		fmt.Println("Cancelled.")
		return ""
	} else {
		checkErr(err)
	}

	return directory
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func addFilePathDict(workFileDict map[string]map[string]bool, fpath string) {
	namePath := fileNameWithoutExtension(fpath)

	_, ok := workFileDict[namePath]
	if !ok {
		dict := make(map[string]bool)
		workFileDict[namePath] = dict
		dict[".lab"] = checkIsFilePath(namePath + ".lab")
		dict[".textgrid"] = checkIsFilePath(namePath + ".textgrid")
	}
}

func removeExistFilePathDict(workFileDict map[string]map[string]bool) {
	for namePath, dict := range workFileDict {
		isFlag := false

		for _, value := range dict {
			if !value {
				isFlag = true
				break
			}
		}

		if !isFlag {
			fmt.Printf("skip... (.textgrid, .lab) already exist. [%v]\n", namePath)
			delete(workFileDict, namePath)
		}
	}
}

func lab2textgrid(namePath string) {
	fmt.Println("Loading...")

	dat, err := os.ReadFile(namePath + ".lab")
	checkErr(err)

	labFile, err := lab.DeserializeLab(dat)
	checkErr(err)

	tg := &textgrid.TextGrid{}

	tg.FileType = "ooTextFile"
	tg.Tiers = true
	tg.Xmin = 0
	tg.Size = 1
	tg.Item = make([]any, 1)
	it := &textgrid.IntervalTier{
		Class: "IntervalTier",
		Name:  "phonemes",
		Xmin:  0,
	}
	tg.Item[0] = it

	if len(labFile.Lines) == 0 {
		tg.Xmax = 0
		it.Xmax = 0
	} else {
		phns := make([]*label.Phoneme, 0)

		for _, line := range labFile.Lines {
			phns = append(phns, &label.Phoneme{
				Text:  line.Text,
				Start: line.Start / float64(_100_NANO_SEC),
				End:   line.End / float64(_100_NANO_SEC),
			})
		}

		if dict.IsValidPhoneDict() {
			pd := dict.CurrentPhoneDict()
			phns = pd.Apply(phns)
		}

		it.Intervals = make([]*textgrid.Interval, 0, len(phns))
		for _, phn := range phns {
			iv := &textgrid.Interval{
				Text: phn.Text,
				Xmin: phn.Start,
				Xmax: phn.End,
			}
			it.Intervals = append(it.Intervals, iv)
		}

		it.Size = len(it.Intervals)
		tg.Xmax = it.Intervals[len(it.Intervals)-1].Xmax
		it.Xmax = tg.Xmax
	}

	str := textgrid.SerializeTextGrid(tg)

	os.WriteFile(namePath+".textgrid", []byte(str), os.FileMode(0644))

	fmt.Println("Save.")
}

func textgrid2lab(namePath string) {
	fmt.Println("Loading...")

	dat, err := os.ReadFile(namePath + ".textgrid")
	checkErr(err)

	tg, err := textgrid.DeserializeTextGrid(dat)
	checkErr(err)

	if len(tg.Item) == 0 {
		fmt.Println("Textgrid is empty.")
	} else {
		for itemIdx, _item := range tg.Item {
			item, ok := _item.(*textgrid.IntervalTier)
			if !ok {
				fmt.Printf("A missing information has occurred. (index: %v)\n", itemIdx)
				continue
			}

			outPath := namePath + "_" + item.Name + ".lab"
			if checkIsFilePath(outPath) {
				fmt.Printf("skip... (%v) already exist.\n", item.Name)
				continue
			}

			labFile := &lab.Lab{}
			labFile.Lines = make([]*lab.Line, 0, len(item.Intervals))

			phns := make([]*label.Phoneme, 0)

			for _, interval := range item.Intervals {
				phns = append(phns, &label.Phoneme{
					Text:  interval.Text,
					Start: interval.Xmin * float64(_100_NANO_SEC),
					End:   interval.Xmax * float64(_100_NANO_SEC),
				})
			}

			if dict.IsValidPhoneDict() {
				pd := dict.CurrentPhoneDict()
				phns = pd.Apply(phns)
			}

			for _, phn := range phns {
				line := &lab.Line{
					Text:  phn.Text,
					Start: phn.Start,
					End:   phn.End,
				}
				labFile.Lines = append(labFile.Lines, line)
			}

			str := lab.SerializeLab(labFile)

			os.WriteFile(outPath, []byte(str), os.FileMode(0644))
		}
	}

	fmt.Println("Save.")
}

func main() {
	workFileDict := make(map[string]map[string]bool)

	getFilePaths := func(dirPath string) []string {
		if dirPath == "" {
			return nil
		}

		pathBuffer := make([]string, 0)

		tmp, err := filepath.Glob(filepath.Join(dirPath, "*.[tT][eE][xX][tT][gG][rR][iI][dD]"))
		checkErr(err)
		pathBuffer = append(pathBuffer, tmp...)

		tmp, err = filepath.Glob(filepath.Join(dirPath, "*", "*.[tT][eE][xX][tT][gG][rR][iI][dD]"))
		checkErr(err)
		pathBuffer = append(pathBuffer, tmp...)

		tmp, err = filepath.Glob(filepath.Join(dirPath, "*.[lL][aA][bB]"))
		checkErr(err)
		pathBuffer = append(pathBuffer, tmp...)

		tmp, err = filepath.Glob(filepath.Join(dirPath, "*", "*.[lL][aA][bB]"))
		checkErr(err)
		pathBuffer = append(pathBuffer, tmp...)

		return pathBuffer
	}

	if len(os.Args) > 1 {
		args := os.Args[1:]
		for _, arg := range args {
			path, err := filepath.Abs(arg)

			if err != nil {
				fmt.Printf("skip... %v [%v]\n", err, arg)
			}

			if checkIsFilePath(path) {
				addFilePathDict(workFileDict, path)
			} else if checkIsDirPath(path) {
				filePaths := getFilePaths(path)

				for _, fpath := range filePaths {
					addFilePathDict(workFileDict, fpath)
				}
			} else {
				fmt.Printf("skip... path does not exist. [%v]\n", path)
			}
		}
	} else {
		dirPath := openDialog()

		filePaths := getFilePaths(dirPath)

		for _, fpath := range filePaths {
			addFilePathDict(workFileDict, fpath)
		}
	}

	removeExistFilePathDict(workFileDict)

	for namePath, dict := range workFileDict {
		fmt.Printf("converting... [%v]\n", namePath)

		isTextgrid := dict[".textgrid"]
		isLab := dict[".lab"]

		if isTextgrid && !isLab {
			fmt.Println("format: [.textgrid -> .lab]")
			textgrid2lab(namePath)
		} else if !isTextgrid && isLab {
			fmt.Println("format: [.lab -> .textgrid]")
			lab2textgrid(namePath)
		}
	}

	fmt.Println("Done.")
	fmt.Scanln()
}
