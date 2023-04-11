package csv

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLineSplit(t *testing.T) {
	tests := []struct {
		line string
		want []string
	}{
		{
			line: `1,"foo",2,"bar",\N,"foo\nbar",foo\nbar,foo\tand\tbar`,
			want: []string{
				"1",
				"foo",
				"2",
				"bar",
				"\\N",
				"foo\nbar",
				"foo\nbar",
				"foo\tand\tbar",
			},
		},
		{
			line: `"TV U600 32\"",Smart TV,32\",,`,
			want: []string{
				"TV U600 32\"",
				"Smart TV",
				"32\"",
				"",
				"",
			},
		},
	}

	for _, test := range tests {
		got, err := LineSplit(test.line)

		require.NoError(t, err, "LineSplit(%q)", test.line)
		assert.Equal(t, test.want, got)
	}
}
