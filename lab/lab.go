package lab

import (
	"fmt"
	"strconv"
	"strings"
)

func DeserializeLab(input []byte) (*Lab, error) {
	var err error
	resultLab := &Lab{}
	resultLab.Lines = make([]*Line, 0)

	cleanStr := strings.ReplaceAll(string(input), "\r\n", "\n")
	strLines := strings.Split(cleanStr, "\n")

	for _, strLine := range strLines {
		strLine = strings.TrimSpace(strLine)
		if strLine == "" {
			continue
		}

		line := &Line{}
		strLineSegs := strings.SplitN(strLine, " ", 3)
		line.Start, err = strconv.ParseFloat(strLineSegs[0], 64)
		if err != nil {
			return nil, err
		}
		line.End, err = strconv.ParseFloat(strLineSegs[1], 64)
		if err != nil {
			return nil, err
		}

		if len(strLineSegs) > 2 {
			line.Text = strLineSegs[2]
		}
		resultLab.Lines = append(resultLab.Lines, line)
	}

	return resultLab, nil
}

func SerializeLab(lab *Lab) string {
	sb := &strings.Builder{}

	for _, line := range lab.Lines {
		sb.WriteString(fmt.Sprintf("%v %v %v\n", int(line.Start), int(line.End), TernaryOperator(line.Text == "", "-", line.Text)))
	}

	return strings.TrimSpace(sb.String())
}
