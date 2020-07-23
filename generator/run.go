// Package generator acts as a prisma generator
package generator

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/prisma/prisma-client-go/binaries"
	"github.com/prisma/prisma-client-go/binaries/bindata"
	"github.com/prisma/prisma-client-go/binaries/platform"
)

const DefaultPackageName = "db"

func addDefaults(input *Root) {
	if input.Generator.Config.Package == "" {
		input.Generator.Config.Package = DefaultPackageName
	}
}

// TODO generate query-engine via go files into user's project dirs
// TODO refactor

var allTargets = []string{
	"darwin",
	"windows",
	// TODO
}

// Run invokes the generator, which builds the templates and writes to the specified output file.
func Run(input *Root) error {
	addDefaults(input)

	targets := input.Generator.BinaryTargets

	if len(targets) == 0 {
		targets = append(targets, "native")
	}

	// copy the query engine to the local repository path
	// TODO refactor
	for _, name := range targets {
		if name == "native" {
			name = platform.BinaryPlatformName()
		}

		// first, ensure they are actually downloaded
		if err := binaries.FetchEngine(binaries.GlobalCacheDir(), "query-engine", name); err != nil {
			return fmt.Errorf("failed fetching binaries: %w", err)
		}

		qe := "prisma-query-engine-" + name

		dir := path.Join(binaries.GlobalCacheDir(), qe)

		input, err := ioutil.ReadFile(dir)
		if err != nil {
			return fmt.Errorf("could not read file %s: %w", name, err)
		}

		dest := "./" + qe
		err = ioutil.WriteFile(dest, input, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not write file to %s: %w", dest, err)
		}
	}

	var buf bytes.Buffer

	ctx := build.Default
	pkg, err := ctx.Import("github.com/prisma/prisma-client-go", ".", build.FindOnly)
	if err != nil {
		return fmt.Errorf("could not get main template asset: %w", err)
	}

	var templates []*template.Template

	templateDir := pkg.Dir + "/generator/templates"
	err = filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".gotpl") {
			tpl, err := template.ParseFiles(path)
			if err != nil {
				return err
			}
			templates = append(templates, tpl.Templates()...)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("could not walk dir %s: %w", templateDir, err)
	}

	// Run header template first
	header, err := template.ParseFiles(templateDir + "/_header.gotpl")
	if err != nil {
		return fmt.Errorf("could not find header template %s: %w", templateDir, err)
	}

	if err := header.Execute(&buf, input); err != nil {
		return fmt.Errorf("could not write header template: %w", err)
	}

	// Then process all remaining templates
	for _, tpl := range templates {
		if strings.Contains(tpl.Name(), "_") {
			continue
		}

		buf.Write([]byte(fmt.Sprintf("// --- template %s ---\n", tpl.Name())))

		if err := tpl.Execute(&buf, input); err != nil {
			return fmt.Errorf("could not write template file %s: %w", tpl.Name(), err)
		}

		if _, err := format.Source(buf.Bytes()); err != nil {
			return fmt.Errorf("could not format source %s from file %s: %w", buf.String(), tpl.Name(), err)
		}
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not format final source: %w", err)
	}

	if strings.HasSuffix(input.Generator.Output, ".go") {
		return fmt.Errorf("generator output should be a directory")
	}

	if err := os.MkdirAll(input.Generator.Output, os.ModePerm); err != nil {
		return fmt.Errorf("could not run MkdirAll on path %s: %w", input.Generator.Output, err)
	}

	// TODO make this configurable
	outFile := path.Join(input.Generator.Output, "db_gen.go")
	if err := ioutil.WriteFile(outFile, formatted, 0644); err != nil {
		return fmt.Errorf("could not write template data to file writer %s: %w", outFile, err)
	}

	if err := generateQueryEngineFiles(targets, input.Generator.Output); err != nil {
		return fmt.Errorf("could not write template data to file writer %s: %w", outFile, err)
	}

	// generate a gitignore into the folder
	var gitignore string
	for _, target := range targets {
		if target == "native" {
			target = platform.BinaryPlatformName()
		}
		gitignore += fmt.Sprintf("query-engine-%s.go\n", target)
	}
	if err := ioutil.WriteFile(path.Join(input.Generator.Output, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("could not write .gitignore: %w", err)
	}

	return nil
}

func generateQueryEngineFiles(binaryTargets []string, outputDir string) error {
	for _, name := range binaryTargets {
		if name == "native" {
			name = platform.BinaryPlatformName()
		}

		enginePath := binaries.GetEnginePath(binaries.GlobalCacheDir(), "query-engine", name)

		pt := name
		if strings.Contains(name, "debian") || strings.Contains(name, "rhel") {
			pt = "linux"
		}
		filename := fmt.Sprintf("query-engine-%s.go", name)
		if err := bindata.WriteFile(name, pt, enginePath, path.Join(outputDir, filename)); err != nil {
			return fmt.Errorf("generate write go file: %w", err)
		}
	}

	return nil
}
