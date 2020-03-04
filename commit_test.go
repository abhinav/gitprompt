package gitprompt

import (
	"reflect"
	"testing"
)

func TestReadCommit(t *testing.T) {
	tests := []struct {
		desc string
		give string
		want Commit
	}{
		{
			desc: "hash only",
			give: "abcdef",
			want: Commit{ShortID: "abcdef"},
		},
		{
			desc: "hash and tags",
			give: "abcdef (HEAD, tag: foo, tag: bar/baz)",
			want: Commit{
				ShortID: "abcdef",
				Tags:    []string{"foo", "bar/baz"},
			},
		},
		{
			desc: "branch",
			give: "abcdef (HEAD -> master, origin/master, tag: foo)",
			want: Commit{
				ShortID: "abcdef",
				Branch:  "master",
				Tags:    []string{"foo"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := ReadCommit(tt.give)
			if err != nil {
				t.Errorf("ReadCommit(%q) = %v", tt.give, err)
				return
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("ReadCommit(%q) = %#v != %#v", tt.give, got, tt.want)
			}
		})
	}
}
