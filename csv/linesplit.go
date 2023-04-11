package csv

import "fmt"

func LineSplit(line string) ([]string, error) {
	var fields []string

	var currentField string
	var inQuotes bool
	for i := 0; i < len(line); i++ {
		c := line[i]

		switch c {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				fields = append(fields, currentField)
				currentField = ""
			}
		case '\\':
			if i+1 >= len(line) {
				return nil, fmt.Errorf("unexpected EOF")
			}
			i++
			c = line[i]
			switch c {
			case 'n':
				currentField += "\n"
			case 't':
				currentField += "\t"
			case ',', '"', '\\':
				currentField += string(c)
			default:
				currentField += "\\" + string(c)
			}
		default:
			currentField += string(c)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("unterminated quoted field")
	}

	fields = append(fields, currentField)

	return fields, nil
}
