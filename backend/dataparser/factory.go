package dataparser

import "fmt"

type adapterConstructor func() DocumentParser

var registry = map[Format]adapterConstructor{}

func RegisterAdapter(format Format, ctor adapterConstructor) {
	registry[format] = ctor
}

func NewParser(format Format) DocumentParser {
	if ctor, ok := registry[format]; ok {
		return ctor()
	}
	panic(fmt.Sprintf("dataparser: no adapter registered for format %q", format))
}

func NewCSVParser() DocumentParser {
	return NewParser(FormatCSV)
}
