package deletion

type Dependency struct{}

func New() (*Dependency, error) {
	return &Dependency{}, nil
}
