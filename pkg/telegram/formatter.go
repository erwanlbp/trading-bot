package telegram

type Formatter struct {
	ColumnSize int
}

var (
	nbCharLine = 45 // Average number of max char display on smartphone line
)

func InitFormatter(nbColumn int) *Formatter {
	return &Formatter{
		ColumnSize: nbCharLine / nbColumn,
	}
}

func (f Formatter) Resize(s string) string {
	if len(s) < f.ColumnSize {
		for i := 1; i < f.ColumnSize-len(s); i++ {
			s += " "
		}
		return s
	}
	return s[0:f.ColumnSize-3] + ".."
}

func (f Formatter) GenerateHeader(headers []string) string {
	header := ""
	headerExceptLast := headers[:len(headers)-1]
	for _, h := range headerExceptLast {
		header += f.Resize(h) + " | "
	}
	header += headers[len(headers)-1:][0]
	return header
}
