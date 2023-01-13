package runner

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Runner is the function signature passed to protogen.Options#Run.
type Runner func(p *protogen.Plugin) error

// New instantiates a Runner that performs visibility checks.
func New() Runner {
	r := &runner{lookup: map[protoreflect.FullName]*rules{}}
	return r.Run
}

type runner struct {
	lookup map[protoreflect.FullName]*rules
}

func (r *runner) Run(p *protogen.Plugin) error {
	var errs []error

	for _, file := range p.Files {
		r.processFile(file)

		if file.Generate {
			errs = append(errs, r.checkFile(file)...)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return VisibilityErrors(errs)
}
