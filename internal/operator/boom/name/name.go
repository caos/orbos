package name

type Application string

func (a Application) String() string {
	return string(a)
}

type Templator string

func (t Templator) String() string {
	return string(t)
}

type Bundle string

func (b Bundle) String() string {
	return string(b)
}

type Version string

func (v Version) String() string {
	return string(v)
}
