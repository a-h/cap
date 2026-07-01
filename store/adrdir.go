package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/a-h/cap/markdown"
	"github.com/a-h/cap/model"
)

// adrDirConfig is the filename written by adr-tools to record the directory that
// holds architectural decision records, for example "doc/adr".
const adrDirConfig = ".adr-dir"

// adrToolsID matches an adr-tools filename, whose identifier is a leading number
// with no prefix, for example "0001" in "0001-record-architecture-decisions.md".
// A "adr-" prefix is also accepted so directories that already follow cap's
// naming continue to load.
var adrToolsID = regexp.MustCompile(`^(?i:adr-)?([0-9]+)(?:-.*)?$`)

// defaultADRDir is the directory adr-tools uses when no .adr-dir config names one.
// cap reads ADRs from it, relative to a directory that has one, so a project that
// follows the convention without a config file is still recognised.
const defaultADRDir = "doc/adr"

// resolveADRDir reports the external directory that holds architectural decision
// records. Each directory from the model root up to the filesystem root is checked
// in turn, matching how adr-tools locates its config from the working directory
// upward. A directory is a match when it contains an adr-tools .adr-dir config file
// naming an existing directory, or, failing that, a conventional doc/adr directory.
// It reports ok=false when no such directory is found.
func resolveADRDir(root string) (dir string, ok bool) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	for _, current := range ancestors(abs) {
		if dir, ok := configuredADRDir(current); ok {
			return dir, true
		}
		if convention := filepath.Join(current, defaultADRDir); isDir(convention) {
			return convention, true
		}
	}
	return "", false
}

// ancestors returns dir followed by each of its parent directories, ending at the
// filesystem root.
func ancestors(dir string) []string {
	var out []string
	for {
		out = append(out, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			return out
		}
		dir = parent
	}
}

// configuredADRDir reads a .adr-dir config file in dir and returns the directory it
// names, resolved relative to dir. It reports ok=false when the config is absent,
// empty, or names a directory that does not exist.
func configuredADRDir(dir string) (adrDir string, ok bool) {
	b, err := os.ReadFile(filepath.Join(dir, adrDirConfig))
	if err != nil {
		return "", false
	}
	named := strings.TrimSpace(string(b))
	if named == "" {
		return "", false
	}
	adrDir = named
	if !filepath.IsAbs(adrDir) {
		adrDir = filepath.Join(dir, adrDir)
	}
	if !isDir(adrDir) {
		return "", false
	}
	return adrDir, true
}

// isDir reports whether path exists and is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// parseADRToolsID derives an ADR identifier from an adr-tools filename. The
// adr-tools naming convention is a zero-padded number followed by a slug, with no
// prefix, for example "0001-record-architecture-decisions.md", which maps to the
// canonical identifier "adr-0001". It reports ok=false when the filename does not
// begin with a number.
func parseADRToolsID(path string) (id model.ID, ok bool) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := adrToolsID.FindStringSubmatch(base)
	if m == nil {
		return "", false
	}
	return model.ID("adr-" + m[1]).Canonical(), true
}

// loadExternalADRs reads every ADR from an adr-tools directory into the model. The
// files there follow adr-tools naming rather than cap's, so identifiers are
// synthesised from the leading number. Loading is tolerant in the same way as the
// rest of the store: problems are recorded but do not stop the load.
func (res *LoadResult) loadExternalADRs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("store: reading %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			continue
		}
		file := filepath.Join(dir, entry.Name())
		res.loadExternalADRFile(file)
	}
	return nil
}

func (res *LoadResult) loadExternalADRFile(file string) {
	id, ok := parseADRToolsID(file)
	if !ok {
		return
	}
	f, err := os.Open(file)
	if err != nil {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("cannot open file: %v", err)})
		return
	}
	defer f.Close()
	doc, err := markdown.Parse(f)
	if err != nil {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("cannot parse document: %v", err)})
		return
	}
	if existingKind, exists := res.Model.Lookup(id); exists {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("duplicate identifier %s (already loaded as %s)", id, existingKind)})
		return
	}
	res.Model.ADRs[id] = model.ADR{ID: id, Title: doc.Title}
	res.Files[id] = file
}
