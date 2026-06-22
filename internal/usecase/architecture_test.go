package usecase

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUsecaseHasNoInfrastructureImports(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve current file")
	}

	usecaseDir := filepath.Dir(thisFile)
	forbidden := []string{
		"github.com/go-chi/chi",
		"github.com/redis/go-redis",
		"github.com/jackc/pgx",
		"github.com/minio/minio-go",
		"net/http",
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, usecaseDir, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse dir: %v", err)
	}

	for _, pkg := range pkgs {
		for fileName, f := range pkg.Files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, "\"")
				for _, bad := range forbidden {
					if strings.Contains(path, bad) {
						t.Fatalf("forbidden usecase import %q in %s", path, fileName)
					}
				}
			}
		}
	}
}
