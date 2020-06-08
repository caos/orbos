package chart

type Chart struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Index   *Index `yaml:"index"`
}

type Index struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
