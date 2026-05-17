package stub

import "github.com/EthanShen10086/voxera-kit/dataparser"

func init() {
	for _, f := range []dataparser.Format{
		dataparser.FormatPDF,
		dataparser.FormatCSV,
		dataparser.FormatXLSX,
		dataparser.FormatHTML,
		dataparser.FormatJSON,
	} {
		format := f
		dataparser.RegisterAdapter(format, func() dataparser.DocumentParser {
			return New()
		})
	}
}
