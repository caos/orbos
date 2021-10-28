package kubernetes

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"strings"
)

var (
	buffer = bytes.Buffer{}

	allGenerators = map[string]genall.Generator{
		"crd":    crd.Generator{},
		"object": deepcopy.Generator{},
	}

	allOutputRules = map[string]genall.OutputRule{
		"dir":       genall.OutputToDirectory(""),
		"none":      genall.OutputToNothing,
		"stdout":    genall.OutputToStdout,
		"buffer":    outputToBuffer{},
		"artifacts": genall.OutputArtifacts{},
	}

	optionsRegistry = &markers.Registry{}
)

func getYamlsFromBuffer() []string {
	yamls := strings.Split(buffer.String(), "---")
	buffer.Reset()
	return yamls
}

func GetCRDs(boilerplatePath, path string) ([]string, error) {
	if err := generateDeepCopy(boilerplatePath, path); err != nil {
		return nil, err
	}

	if err := generateCrd(path); err != nil {
		return nil, err
	}

	return getYamlsFromBuffer(), nil
}

func WriteCRDs(boilerplatePath, path, destinationFolder string) error {
	if err := generateDeepCopy(boilerplatePath, path); err != nil {
		return err
	}

	return generateCrdToFolder(path, destinationFolder)
}

func ApplyCRDs(boilerplatePath, path string, k8sClient ClientInt) error {
	if err := generateDeepCopy(boilerplatePath, path); err != nil {
		return err
	}

	if err := generateCrd(path); err != nil {
		return err
	}

	for _, crd := range getYamlsFromBuffer() {
		if crd == "" || crd == "\n" {
			continue
		}

		//crdDefinition := apixv1beta1.CustomResourceDefinition{}
		crdDefinition := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(crd), &crdDefinition.Object); err != nil {
			return err
		}

		if err := k8sClient.ApplyCRDResource(
			crdDefinition,
		); err != nil {
			return err
		}
	}
	return nil
}

type outputToBuffer struct{}

func (o outputToBuffer) Open(_ *loader.Package, itemPath string) (io.WriteCloser, error) {
	return nopCloser{&buffer}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}

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

func generate(rawOpts []string) error {
	rt, err := genall.FromOptions(optionsRegistry, rawOpts)
	if err != nil {
		panic(err)
	}
	if len(rt.Generators) == 0 {
		return fmt.Errorf("no generators specified")
	}

	if hadErrs := rt.Run(); hadErrs {
		return fmt.Errorf("not all generators ran successfully")
	}

	return nil
}

func generateDeepCopy(boilerplatePath, path string) error {
	return generate([]string{
		"object:headerFile=\"" + boilerplatePath + "\"",
		"paths=\"" + path + "\"",
	})
}

func generateCrd(path string) error {
	return generate([]string{
		"crd:trivialVersions=true",
		"crd",
		"paths=\"" + path + "\"",
		"output:crd:buffer",
	})
}

func generateCrdToFolder(path, folder string) error {
	return generate([]string{
		"crd:trivialVersions=true",
		"crd",
		"paths=\"" + path + "\"",
		"output:crd:artifacts:config=\"" + folder + "\"",
	})
}
