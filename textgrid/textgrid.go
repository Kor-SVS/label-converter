// Package praatgo provides tools to interpret and process
// Praat output and input objects as defined in
// https://www.fon.hum.uva.nl/praat/manual/Types_of_objects
package textgrid

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// DeserializeTextGrid recurses the content of a TextGrid file
// and returns its values into tg
func DeserializeTextGrid(input []byte) (tg *TextGrid, err error) {
	content := bytes.ReplaceAll(input, []byte("\x00"), []byte{}) // praat usually writes null bytes between characters
	pattern := regexp.MustCompile(`"(.*?)"`)
	headers := pattern.FindAllSubmatch(content, 2)
	if !bytes.Equal(headers[1][1], []byte("TextGrid")) {
		return tg, errors.New("not a TextGrid file")
	}
	tg = &TextGrid{}
	tg.FileType = string(headers[0][1])
	cursor := pattern.FindAllSubmatchIndex(content, 2)[1][1]
	content = content[cursor:]
	pattern = regexp.MustCompile(`\s[0-9\.]+|"(.*?)"|<exists>|<absent>`)
	data := pattern.FindAll(content, -1)
	tg.Xmin, err = parseNumber(data[0])
	if err != nil {
		return
	}
	tg.Xmax, err = parseNumber(data[1])
	if err != nil {
		return
	}
	tg.Tiers, err = parseBool(data[2])
	if err != nil {
		return
	}
	tg.Size, err = parseIndex(data[3])
	if err != nil {
		return
	}
	j := 4
	for i1 := 0; i1 < tg.Size; i1++ {
		switch string(data[j]) {
		case `"IntervalTier"`:
			tier, err := parseIntervalTier(data[j+1:])
			if err != nil {
				return tg, err
			}
			j += 5 + 3*len(tier.Intervals)
			tg.Item = append(tg.Item, tier)
		default:
			return tg, errors.New("bad format")
		}
	}
	return
}

func SerializeTextGrid(tg *TextGrid) string {
	sb := &strings.Builder{}

	sb.WriteString(fmt.Sprintf("%v = \"%v\"\n", getStructTag(tg, "FileType", "json"), tg.FileType))
	sb.WriteString("Object class = \"TextGrid\"\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%v = %v\n", getStructTag(tg, "Xmin", "json"), tg.Xmin))
	sb.WriteString(fmt.Sprintf("%v = %v\n", getStructTag(tg, "Xmax", "json"), tg.Xmax))
	sb.WriteString(fmt.Sprintf("%v? %v\n", getStructTag(tg, "Tiers", "json"), TernaryOperator(tg.Tiers, "<exists>", "<absent>")))
	sb.WriteString(fmt.Sprintf("%v = %v\n", getStructTag(tg, "Size", "json"), tg.Size))

	if tg.Size > 0 {
		sb.WriteString("item []: \n")
		for itemIdx, _item := range tg.Item {
			sb.WriteString(fmt.Sprintf("    item [%v]:\n", itemIdx+1))

			switch item := _item.(type) {
			case *IntervalTier:
				sb.WriteString(fmt.Sprintf("        %v = \"%v\"\n", getStructTag(item, "Class", "json"), item.Class))
				sb.WriteString(fmt.Sprintf("        %v = \"%v\"\n", getStructTag(item, "Name", "json"), item.Name))
				sb.WriteString(fmt.Sprintf("        %v = %v\n", getStructTag(item, "Xmin", "json"), item.Xmin))
				sb.WriteString(fmt.Sprintf("        %v = %v\n", getStructTag(item, "Xmax", "json"), item.Xmax))
				sb.WriteString(fmt.Sprintf("        intervals: %v = %v\n", getStructTag(item, "Size", "json"), item.Size))

				for intervalIdx, interval := range item.Intervals {
					sb.WriteString(fmt.Sprintf("        intervals [%v]:\n", intervalIdx+1))
					sb.WriteString(fmt.Sprintf("            %v = %v\n", getStructTag(interval, "Xmin", "json"), interval.Xmin))
					sb.WriteString(fmt.Sprintf("            %v = %v\n", getStructTag(interval, "Xmax", "json"), interval.Xmax))
					sb.WriteString(fmt.Sprintf("            %v = \"%v\"\n", getStructTag(interval, "Text", "json"), interval.Text))
				}

			case *TextTier:
				sb.WriteString(fmt.Sprintf("        %v = \"%v\"\n", getStructTag(item, "Class", "json"), item.Class))
				sb.WriteString(fmt.Sprintf("        %v = \"%v\"\n", getStructTag(item, "Name", "json"), item.Name))
				sb.WriteString(fmt.Sprintf("        %v = %v\n", getStructTag(item, "Xmin", "json"), item.Xmin))
				sb.WriteString(fmt.Sprintf("        %v = %v\n", getStructTag(item, "Xmax", "json"), item.Xmax))
				sb.WriteString(fmt.Sprintf("        points: %v = %v\n", getStructTag(item, "Size", "json"), item.Size))

				for pointIdx, point := range item.Points {
					sb.WriteString(fmt.Sprintf("        points [%v]:\n", pointIdx+1))
					sb.WriteString(fmt.Sprintf("            %v = %v\n", getStructTag(point, "Number", "json"), point.Number))
					sb.WriteString(fmt.Sprintf("            %v = \"%v\"\n", getStructTag(point, "Mark", "json"), point.Mark))
				}
			}
		}
	}

	return sb.String()
}
