package parser

import (
	"testing"
	"os"
	"path/filepath"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)


func TestParseTree(t *testing.T) {
	tmp := createDummyLibrary(t)
	logger := slog.Default()

	root, err := ParseTree(tmp, nil, 0, logger)
	if err != nil {
		t.Errorf("ParseTree returns error %v", err)
	}

	if root.Parent != nil {
		t.Errorf("ParseTree root not nil, want nil")
	}
	
	if len(root.Children) != 2 {
		t.Errorf("ParseTree root children len %v, want 2", len(root.Children))
	}

	depth := getDepth(root.Children[0], 1) 
	if depth != 2 {
		t.Errorf("Expected depth of 2 for tempdir/dir/file.txt, got %v", depth) 
	}

	depth = root.Children[0].Children[0].Depth
	if depth != 2 {
		t.Errorf("Expected calculated Depth for tempdir/dir/file.txt by ParseTree = %v, got %v", 2, depth) 
	}

	depth = getDepth(root.Children[1], 1)
	if depth != 4 {
		t.Errorf("Expected depth of 4 for tempdir/parent/child/subchild/file.txt, got %v", depth) 
	}

	depth = root.Children[1].Children[0].Children[0].Children[0].Depth
	if depth != 4 {
		t.Errorf("Expected calculated Depth for tempdir/parent/child/subchild/file.txt by ParseTree = %v, got %v", 4, depth) 
	}

	_, err = ParseTree("/nonexistent/path", nil, 0, slog.Default())
    if err == nil {
        t.Error("expected error for nonexistent path")
    }
	
}

func getDepth(entry *metadata.Entry, depth int) int {
	if len(entry.Children) == 0 {
		return depth	
	}

	max := depth
	for _, child := range entry.Children {
		childDepth := getDepth(child, depth + 1)	

		if childDepth > max {
			max = childDepth
		}
	}

	return max
}

//Makes two dummy direcotries inside temporary directory
// temp dir -> dir -> file.txt
// temp dir -> parent -> child -> subchild -> file.txt
func createDummyLibrary(t *testing.T) string {
	dir := t.TempDir()

	dummyDir := filepath.Join(dir, "dir")
	err := os.MkdirAll(dummyDir, 0755)
	if err != nil {
		t.Fatalf("Unable to create dummy dir %v, error %v\n", dummyDir, err)
	}

	dummyFile := filepath.Join(dummyDir, "file.txt")
	_, err = os.Create(dummyFile)
	if err != nil {
		t.Fatalf("Unable to create dummy file %v, error %v", dummyFile, err)
	}


	dummyDir = filepath.Join(dir, "parent", "child", "subchild")
	err = os.MkdirAll(dummyDir, 0755)
	if err != nil {
		t.Fatalf("Unable to create dummy dir %v, error %v\n", dummyDir, err)
	}

	dummyFile = filepath.Join(dummyDir, "file.txt")
	_, err = os.Create(dummyFile)
	if err != nil {
		t.Fatalf("Unable to create dummy file %v, error %v", dummyFile, err)
	}

	invalidFile := filepath.Join(dummyDir, "invalid.jpg")
	_, err = os.Create(invalidFile)
	if err != nil {
		t.Fatalf("Unable to create invalid file %v, error %v", invalidFile, err)
	}
	invalidFile = filepath.Join(dummyDir, "invalid.txt")
	_, err = os.Create(invalidFile)
	if err != nil {
		t.Fatalf("Unable to create invalid file %v, error %v", invalidFile, err)
	}


	return dir
}
