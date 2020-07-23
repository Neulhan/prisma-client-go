package unpack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"time"
)

// TODO check checksum after expanding file

//noinspection GoUnusedExportedFunction
func Unpack(data []byte, name string) {
	file := fmt.Sprintf("prisma-%s-%s", name, runtime.GOOS)

	// TODO check if dev env/dev binary in ~/.prisma
	// TODO check if engine in local dir OR env var

	start := time.Now()
	dir := path.Join(".", file)
	// dir := path.Join(os.TempDir(), file)
	if err := ioutil.WriteFile(dir, data, os.ModePerm); err != nil {
		panic(fmt.Errorf("unpack write file: %w", err))
	}
	log.Printf("unpacked at %s in %s", dir, time.Since(start))
}
