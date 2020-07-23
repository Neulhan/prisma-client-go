package bindata

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// TODO go fmt files after creation

// WriteFile writes the release code file.
func WriteFile(name, platform, from, to string) error {
	f, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("generate open go file: %w", err)
	}

	defer f.Close()

	if err := writeHeader(f, name, platform); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if err := writeAsset(f, from); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}

	return nil
}

// writeHeader writes output file headers.
// This targets release builds.
func writeHeader(w io.Writer, name, platform string) error {
	var excludes string
	if platform == "linux-musl" {
		platform = "linux"
		// TODO dynamically construct these with allTargets in run.go
		// TODO only include these for engines, not for the CLI
		excludes = `
// +build !debian_openssl_1_0_x
// +build !debian_openssl_1_1_x
// +build !rhel_openssl_1_0_x
// +build !rhel_openssl_1_1_x
		`
	}

	platform = strings.Replace(platform, "-", "_", -1)
	platform = strings.Replace(platform, ".", "_", -1)

	pkg := "db"

	_, err := fmt.Fprintf(w, `
// +build %s
// +build !prisma_ignore

%s

package %s

import (
	"github.com/prisma/prisma-client-go/binaries/unpack"
)

func init() {
	unpack.Unpack(data, "%s-%s")
}

`, platform, excludes, pkg, name, platform)
	return err
}

// writeAsset write a release entry for the given asset.
// A release entry is a function which embeds and returns
// the file's byte content.
func writeAsset(w io.Writer, file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}

	defer fd.Close()

	h := sha256.New()
	tr := io.TeeReader(fd, h)
	if err := uncompressedMemcopy(w, tr); err != nil {
		return err
	}
	var digest [sha256.Size]byte
	copy(digest[:], h.Sum(nil))
	return assetMetadata(w, file, digest)
}

func uncompressedMemcopy(w io.Writer, r io.Reader) error {
	if _, err := fmt.Fprintf(w, `var data = []byte(`); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%+q", b)

	if _, err := fmt.Fprintf(w, `)`); err != nil {
		return err
	}
	return nil
}

func assetMetadata(w io.Writer, file string, digest [sha256.Size]byte) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}

	// TODO
	mode := uint(fi.Mode())
	modTime := fi.ModTime().Unix()
	size := fi.Size()
	_, err = fmt.Fprintf(w, `
//func Data() (*asset, error) {
//	info := bindataFileInfo{name: "bytes", size: %d, mode: os.FileMode(%#o), modTime: time.Unix(%d, 0)}
//	a := &asset{bytes: data, info: info, digest: %#v}
//	return a, nil
//}

`, size, mode, modTime, digest)
	return err
}
