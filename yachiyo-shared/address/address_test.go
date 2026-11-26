package address

import "testing"

func TestAddressPath(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "wsjson://testclient", want: ""},
		{raw: "wsjson://testclient/", want: ""},
		{raw: "wsjson://testclient/group/168", want: "group/168"},
		{raw: "wsjson://testclient/private/1/", want: "private/1"},
	}

	for _, tc := range tests {
		got, err := New(tc.raw).Path()
		if err != nil {
			t.Fatalf("Path(%q) unexpected error: %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("Path(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}
