package jobmanager

import "fmt"

type Argument struct {
	Name  string
	Value string
}

func (a *Argument) String() string {
	return fmt.Sprintf("-%s %s", a.Name, a.Value)
}

type Arguments []Argument

func (a *Arguments) String() []string {
	args := make([]string, 0)
	for _, arg := range *a {
		args = append(args, arg.String())
	}
	return args
}
