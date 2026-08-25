package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/skmohammadali786/gogo/internal/compiler"
	"github.com/skmohammadali786/gogo/internal/config"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
)

const version = "0.1.0-dev"

func main() {
	jsonOut := flag.Bool("json", false, "print diagnostics as JSON")
	locale := flag.String("locale", "en", "diagnostic language: en, bn, or hi")
	grammarLanguage := flag.String("grammar", "", "override project language/grammar: en, bn, or hi")
	strictness := flag.String("strictness", "", "override project strictness: standard, strict, or permissive")
	target := flag.String("target", "", "override project target (currently ast)")
	configPath := flag.String("config", "", "project configuration path (default: discover gogo.json)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Printf("GOGO %s\n", version)
		fmt.Println("The GOGO compiler is under active development.")
		return
	}

	raw := config.Raw{}
	usedConfig := ""
	if *configPath != "" {
		loaded, err := config.Load(*configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		raw, usedConfig = loaded, *configPath
	} else if flag.NArg() > 0 {
		loaded, path, ok, err := config.LoadDiscover(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if ok {
			raw, usedConfig = loaded, path
		}
	}
	resolved, configDiagnostics := config.Resolve(raw, config.Overrides{Language: *grammarLanguage, Strictness: *strictness, Target: *target, ConfigPath: usedConfig})
	s := compiler.NewSession(compiler.WithResolvedConfig(resolved))
	for _, d := range configDiagnostics {
		d.FilePath = usedConfig
		s.Diagnostics.Add(d)
	}
	for _, path := range flag.Args() {
		data, err := os.ReadFile(path)
		if err != nil {
			s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G0005", Message: "I could not read this source file.", Hint: err.Error(), FilePath: path})
			continue
		}
		id := s.AddFile(path, string(data))
		if id != 0 {
			s.ParseFile(id)
		}
	}
	r := diagnostics.Renderer{Files: s.Files, Locale: diagnostics.Locale(*locale)}
	if *jsonOut {
		data, err := r.JSON(s.Diagnostics.All())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	} else {
		fmt.Print(r.Text(s.Diagnostics.All()))
	}
	if s.HasErrors() {
		os.Exit(1)
	}
}
