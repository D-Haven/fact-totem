package version

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintsVersionToWriter(t *testing.T) {
	var b bytes.Buffer

	if err := Print(&b); err != nil {
		t.Fatalf("version.Print() gave error: %s", err)
	}

	var v version
	if err := json.Unmarshal(b.Bytes(), &v); err != nil {
		t.Fatalf("Invalid version json: %s", err)
	}

	if v.Release != Release {
		t.Fatalf("Release does not match: %s != %s", Release, v.Release)
	}

	if v.BuildTime != BuildTime {
		t.Fatalf("Build Time does not match: %s != %s", BuildTime, v.BuildTime)
	}

	if v.Commit != Commit {
		t.Fatalf("Commit does not match: %s != %s", Commit, v.Commit)
	}
}
