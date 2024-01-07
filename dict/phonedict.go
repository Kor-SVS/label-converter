package dict

import (
	"errors"
	"fmt"
	"labelconverter/label"
	"strings"

	"github.com/fatih/color"
)

const (
	PhoneDict_SEP = "="
)

type Phones []*label.Phoneme

type PhoneDict struct {
	ruleIndex             []string          // 규칙의 순서
	rulebook              map[string]string // 규칙
	ruleKeySpaceCountDict map[string]int    // 규칙 Key의 Space 개수
}

func NewPhoneDict(rawdata []byte) (pd *PhoneDict, err error) {
	rawtxt := strings.TrimSpace(string(rawdata))
	if rawtxt == "" {
		return nil, errors.New("rule data is empty")
	}

	rawtxt = strings.ReplaceAll(rawtxt, "\r\n", "\n")
	strLines := strings.Split(rawtxt, "\n")
	pd = &PhoneDict{
		ruleIndex:             make([]string, 0, len(strLines)),
		rulebook:              make(map[string]string),
		ruleKeySpaceCountDict: make(map[string]int),
	}

	for _, strLine := range strLines {
		strLine = strings.TrimSpace(strLine)
		if strLine == "" {
			continue
		}

		if !strings.Contains(strLine, PhoneDict_SEP) {
			return nil, fmt.Errorf("rule parsing error [%v]", strLine)
		}

		segs := strings.SplitN(strLine, PhoneDict_SEP, 2)

		key := strings.TrimSpace(segs[0])
		value := strings.TrimSpace(segs[1])

		pd.ruleIndex = append(pd.ruleIndex, key)
		pd.rulebook[key] = value
		pd.ruleKeySpaceCountDict[key] = strings.Count(key, " ")
	}

	return
}

func (pd *PhoneDict) Count() int {
	return len(pd.ruleIndex)
}

func (pd *PhoneDict) getJoinedPhnText(phones Phones) string {
	buf := make([]string, 0, len(phones))
	for _, phn := range phones {
		buf = append(buf, phn.Text)
	}
	return strings.Join(buf, " ")
}

func (pd *PhoneDict) Apply(phones Phones) (rPhones Phones) {
	printConv := color.New(color.FgHiYellow).PrintfFunc()
	printPass := color.New(color.FgGreen).PrintfFunc()
	rPhones = make(Phones, 0, len(phones))

	for phnIdx := 0; phnIdx < len(phones); phnIdx++ {
		isRuleUsed := false
		phn := phones[phnIdx]

		for _, ruleKey := range pd.ruleIndex {
			ruleKeySpaceCount := pd.ruleKeySpaceCountDict[ruleKey]

			if ruleKeySpaceCount == 0 {
				if ruleKey == phn.Text {
					vPhn := pd.rulebook[ruleKey]
					rPhones = append(rPhones, &label.Phoneme{
						Text:  vPhn,
						Start: phn.Start,
						End:   phn.End,
					})
					isRuleUsed = true
					printConv("[PhoneDict Conv] %v -> %v\n", ruleKey, vPhn)
					break
				}
			} else {
				if phnIdx+ruleKeySpaceCount+1 >= len(phones) {
					continue
				}

				rangePhones := phones[phnIdx : phnIdx+ruleKeySpaceCount+1]
				joinedPhnText := pd.getJoinedPhnText(rangePhones)
				if ruleKey == joinedPhnText {
					vPhn := pd.rulebook[ruleKey]
					rPhones = append(rPhones, &label.Phoneme{
						Text:  vPhn,
						Start: rangePhones[0].Start,
						End:   rangePhones[len(rangePhones)-1].End,
					})
					phnIdx += ruleKeySpaceCount
					isRuleUsed = true
					printConv("[PhoneDict Conv] %v -> %v\n", ruleKey, vPhn)
					break
				}
			}
		}

		if !isRuleUsed {
			rPhones = append(rPhones, phn)
			printPass("[PhoneDict Pass] %v\n", phn.Text)
		}
	}

	if len(phones) != len(rPhones) {
		printConv("[PhoneDict Conv] Finish (Count: %v -> %v)\n", len(phones), len(rPhones))
	}

	return
}
