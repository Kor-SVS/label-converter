package lab

import "labelconverter/label"

type Line label.Phoneme

type Lab struct {
	Lines []*Line
}
