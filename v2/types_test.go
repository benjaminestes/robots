package robots

import "testing"

func TestRobotsPath(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{"", "/"},
		{"?q=123", "/?q=123"},
		{"/page.html", "/page.html"},
		{"/page.html#fragment", "/page.html"},
		{"/page.html?q=123", "/page.html?q=123"},
		{"/page.html?q=123#fragment", "/page.html?q=123"},
		{"http://www.example.com/page.html", "/page.html"},
		{"http://www.example.com/page.html#fragment", "/page.html"},
		{"http://www.example.com/page.html?q=123", "/page.html?q=123"},
		{"http://www.example.com/page.html?q=123#fragment", "/page.html?q=123"},
	}

	for _, test := range tests {
		got, ok := robotsPath(test.input)
		if !ok {
			t.Errorf("couldn't parse %q", test.input)
		}
		if got != test.want {
			t.Errorf("crawlable part of %q is %q, got %q",
				test.input, test.want, got)
		}
	}
}
