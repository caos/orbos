package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	crdFolder = "crds"
)

func main() {
	var basePath string
	var boilerplatePath string
	flag.StringVar(&basePath, "basepath", "./artifacts", "The local path where the base folder should be")
	flag.StringVar(&boilerplatePath, "boilerplatepath", "./hack/boilerplate.go.txt", "The local path where the boilerplate text file lies")
	flag.Parse()

	generate([]string{
		"object:headerFile=\"" + boilerplatePath + "\"",
		"paths=\"./...\"",
	})

	generate([]string{
		"crd:trivialVersions=true",
		"crd",
		"paths=\"./...\"",
		"output:crd:artifacts:config=" + filepath.Join(basePath, crdFolder),
	})
}

func generate(rawOpts []string) {
	rt, err := genall.FromOptions(optionsRegistry, rawOpts)
	if err != nil {
		panic(err)
	}
	if len(rt.Generators) == 0 {
		panic(fmt.Errorf("no generators specified"))
	}

	if hadErrs := rt.Run(); hadErrs {
		panic(fmt.Errorf("not all generators ran successfully"))
	}
}

var (
	// allGenerators maintains the list of all known generators, giving
	// them names for use on the command line.
	// each turns into a command line option,
	// and has options for output forms.
	allGenerators = map[string]genall.Generator{
		"crd":    crd.Generator{},
		"object": deepcopy.Generator{},
	}

	// allOutputRules defines the list of all known output rules, giving
	// them names for use on the command line.
	// Each output rule turns into two command line options:
	// - output:<generator>:<form> (per-generator output)
	// - output:<form> (default output)
	allOutputRules = map[string]genall.OutputRule{
		"dir":       genall.OutputToDirectory(""),
		"none":      genall.OutputToNothing,
		"stdout":    genall.OutputToStdout,
		"artifacts": genall.OutputArtifacts{},
	}

	// optionsRegistry contains all the marker definitions used to process command line options
	optionsRegistry = &markers.Registry{}
)

func init() {
	for genName, gen := range allGenerators {
		// make the generator options marker itself
		defn := markers.Must(markers.MakeDefinition(genName, markers.DescribesPackage, gen))
		if err := optionsRegistry.Register(defn); err != nil {
			panic(err)
		}
		if helpGiver, hasHelp := gen.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(defn, help)
			}
		}

		// make per-generation output rule markers
		for ruleName, rule := range allOutputRules {
			ruleMarker := markers.Must(markers.MakeDefinition(fmt.Sprintf("output:%s:%s", genName, ruleName), markers.DescribesPackage, rule))
			if err := optionsRegistry.Register(ruleMarker); err != nil {
				panic(err)
			}
			if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
				if help := helpGiver.Help(); help != nil {
					optionsRegistry.AddHelp(ruleMarker, help)
				}
			}
		}
	}

	// make "default output" output rule markers
	for ruleName, rule := range allOutputRules {
		ruleMarker := markers.Must(markers.MakeDefinition("output:"+ruleName, markers.DescribesPackage, rule))
		if err := optionsRegistry.Register(ruleMarker); err != nil {
			panic(err)
		}
		if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(ruleMarker, help)
			}
		}
	}

	// add in the common options markers
	if err := genall.RegisterOptionsMarkers(optionsRegistry); err != nil {
		panic(err)
	}
}
