package dataparser

import "fmt"

type adapterConstructor func() DocumentParser

var registry = map[Format]adapterConstructor{}

// RegisterAdapter registers a parser constructor for the given format.
func RegisterAdapter(format Format, ctor adapterConstructor) {
	registry[format] = ctor
}

// NewParser returns a DocumentParser for the given format, panicking if none is registered.
func NewParser(format Format) DocumentParser {
	if ctor, ok := registry[format]; ok {
		return ctor()
	}
	panic(fmt.Sprintf("dataparser: no adapter registered for format %q", format))
}

// NewCSVParser is a convenience constructor for a CSV document parser.
func NewCSVParser() DocumentParser {
	return NewParser(FormatCSV)
}
