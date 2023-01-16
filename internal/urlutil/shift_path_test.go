package urlutil_test

import (
	"testing"

	"source.toby3d.me/toby3d/auth/internal/urlutil"
)

func TestShiftPath(t *testing.T) {
	t.Parallel()

	for in, out := range map[string][2]string{
		"/":         {"", "/"},
		"/foo":      {"foo", "/"},
		"/foo/":     {"foo", "/"},
		"/foo/bar":  {"foo", "/bar"},
		"/foo/bar/": {"foo", "/bar"},
	} {
		in, out := in, out

		t.Run(in, func(t *testing.T) {
			t.Parallel()

			head, path := urlutil.ShiftPath(in)

			if out[0] != head || out[1] != path {
				t.Errorf("ShiftPath(%s) = '%s', '%s', want '%s', '%s'", in, head, path, out[0], out[1])
			}
		})
	}
}
