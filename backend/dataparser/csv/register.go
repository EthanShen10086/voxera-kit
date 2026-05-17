package csv

import "github.com/EthanShen10086/voxera-kit/dataparser"

func init() {
	dataparser.RegisterAdapter(dataparser.FormatCSV, func() dataparser.DocumentParser {
		return New()
	})
}
