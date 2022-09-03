package scraping

import (
	"testing"
)

func TestDehumanitizeViewNumber(t *testing.T) {
	t.Parallel()
	type table struct {
		input string
		want  int
	}

	tables := []table{
		{input: "1K", want: 1000},
		{input: "1.5K", want: 1500},
		{input: "12.5K", want: 12500},
		{input: "4.91M", want: 4910000},
		{input: "0M", want: 0},
		{input: "0", want: 0},
		{input: "55", want: 55},
	}

	for _, test := range tables {
		result := dehumanizeViewNumber(test.input)

		if result != test.want {
			t.Errorf("Dehumanized incorrectly on input %s, got: %d, want: %d.", test.input, result, test.want)
		}
	}
}

func TestConvertTgPublicationToAbstractPublication(t *testing.T) {
	t.Parallel()
	tm := telegramPublication{"tg/msgid", "10.5K", "2022-08-29T11:41:26+00:00"}
	generalized := tm.generalize()

	if generalized.PostedAt.String() != "2022-08-29 11:41:26 +0000 UTC" {
		t.Errorf(
			"PostedAt generalized incorrectly, got %s, expected: %s",
			generalized.PostedAt.String(),
			"2022-08-29 11:41:26 +0000 UTC",
		)
	}
}
