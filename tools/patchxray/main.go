// Command patchxray prepares the xray-core that keqrnel is built against: the
// revision pinned in go.mod, downloaded through the module proxy, with
// patches/xray-*.patch applied. It writes a module file that points keqrnel at
// that tree, so a build differs from a plain one by the patches only.
//
//	go run ./tools/patchxray
//	go build -modfile go.patched.mod -tags with_gvisor ./cmd/keqrnel
//	go test -modfile go.patched.mod ./...
//
// Run it from the module root. -extra adds patch directories applied after
// keqrnel's own, for builds that need more (keqdroid's libxray.so). git must be
// on PATH: it applies the patches and refuses one that no longer fits.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const xrayModule = "github.com/xtls/xray-core"

func main() {
	src := flag.String("src", ".build/xray-core", "where to put the patched tree, relative to the module root")
	modfile := flag.String("modfile", "go.patched.mod", "module file to write in the module root")
	extra := flag.String("extra", "", "comma separated directories with more xray-*.patch")
	flag.Parse()

	if err := run(*src, *modfile, *extra); err != nil {
		fmt.Fprintln(os.Stderr, "patchxray:", err)
		os.Exit(1)
	}
}

func run(src, modfile, extra string) error {
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("run from the keqrnel module root: %w", err)
	}
	// relative paths keep the user name out of the binary's build info, and
	// a modfile next to go.mod resolves them the same way go.mod would
	if filepath.IsAbs(src) || strings.HasPrefix(filepath.Clean(src), "..") {
		return fmt.Errorf("-src must be inside the module root: %s", src)
	}
	if strings.ContainsAny(modfile, `/\`) || !strings.HasSuffix(modfile, ".mod") || modfile == "go.mod" {
		return fmt.Errorf("-modfile must be a *.mod file name in the module root: %s", modfile)
	}

	mod, err := download()
	if err != nil {
		return err
	}
	fmt.Printf("%s %s\n", xrayModule, mod.Version)

	if err := os.RemoveAll(src); err != nil {
		return err
	}
	if err := copyTree(mod.Dir, src); err != nil {
		return err
	}

	patches, err := collect("patches", extra)
	if err != nil {
		return err
	}
	for _, p := range patches {
		fmt.Println("applying", p)
		cmd := exec.Command("git", "apply", "-p1", "--whitespace=nowarn",
			"--directory="+filepath.ToSlash(src), p)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s does not apply to xray-core %s, rebase it: %w", p, mod.Version, err)
		}
	}

	goMod, err := os.ReadFile("go.mod")
	if err != nil {
		return err
	}
	text := strings.TrimRight(string(goMod), "\n") +
		"\n\nreplace " + xrayModule + " => ./" + filepath.ToSlash(filepath.Clean(src)) + "\n"
	if err := os.WriteFile(modfile, []byte(text), 0o644); err != nil {
		return err
	}
	sum, err := os.ReadFile("go.sum")
	if err != nil {
		return err
	}
	if err := os.WriteFile(strings.TrimSuffix(modfile, ".mod")+".sum", sum, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s, build with -modfile %s\n", modfile, modfile)
	return nil
}

type module struct {
	Version string
	Dir     string
	Error   string
}

// The version comes from go.mod: `go mod download` without @version takes the
// required one and fails if go.sum does not vouch for it.
func download() (*module, error) {
	out, err := exec.Command("go", "mod", "download", "-json", xrayModule).Output()
	if err != nil {
		return nil, fmt.Errorf("go mod download %s: %w", xrayModule, err)
	}
	var m module
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	if m.Error != "" || m.Dir == "" {
		return nil, fmt.Errorf("go mod download %s: %s", xrayModule, m.Error)
	}
	return &m, nil
}

func collect(own, extra string) ([]string, error) {
	dirs := []string{own}
	for _, d := range strings.Split(extra, ",") {
		if d = strings.TrimSpace(d); d != "" {
			dirs = append(dirs, d)
		}
	}
	var all []string
	for _, d := range dirs {
		found, err := filepath.Glob(filepath.Join(d, "xray-*.patch"))
		if err != nil {
			return nil, err
		}
		sort.Strings(found)
		all = append(all, found...)
	}
	return all, nil
}

// The module cache is read-only, so the copy gets ordinary permissions.
func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}
