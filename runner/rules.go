package runner

import (
	"fmt"
	"log"
	"strings"

	"github.com/rodaine/protoc-gen-visibility/visibility/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func warn(v ...any) {
	v = append([]any{"[WARNING]"}, v...)
	log.Println(v...)
}

type rules struct {
	mode            visibility.Mode
	packagePrefixes map[protoreflect.FullName]struct{}
	packages        map[protoreflect.FullName]struct{}
}

func processRules(desc protoreflect.Descriptor, spec *visibility.Specification) *rules {
	out := &rules{
		mode:            spec.GetMode(),
		packagePrefixes: map[protoreflect.FullName]struct{}{},
		packages:        map[protoreflect.FullName]struct{}{},
	}

	for _, pkg := range spec.GetPackages() {
		name := protoreflect.FullName(strings.TrimSuffix(pkg, ".*"))
		if !name.IsValid() {
			warn(pkg, "is not a valid package name in visibility rules for", desc.FullName())
		}

		if strings.HasSuffix(pkg, ".*") {
			out.packagePrefixes[protoreflect.FullName(strings.TrimSuffix(pkg, "*"))] = struct{}{}
		}

		out.packages[name] = struct{}{}
	}

	return out
}

//nolint:cyclop
func (r *rules) checkUsage(user protoreflect.Descriptor, entity protoreflect.Descriptor) error {
	userFile, userPkg := user.ParentFile(), user.ParentFile().Package()
	entityFile, entityPkg := entity.ParentFile(), entity.ParentFile().Package()

	switch r.mode {
	case visibility.Mode_MODE_FILE:
		if userFile == entityFile {
			return nil
		}
	case visibility.Mode_MODE_SUBPACKAGES:
		if strings.HasPrefix(string(userPkg), string(entityPkg)+".") {
			return nil
		}

		fallthrough // to support the current package as well
	case visibility.Mode_MODE_PACKAGE:
		if userPkg == entityPkg {
			return nil
		}
	case
		visibility.Mode_MODE_UNSPECIFIED,
		visibility.Mode_MODE_PUBLIC:
		return nil
	default:
		warn(entity.FullName(), "visibility mode is unknown:", r.mode)
	}

	for pkg := range r.packages {
		if userPkg == pkg {
			return nil
		}
	}

	for pkgPrefix := range r.packagePrefixes {
		if strings.HasPrefix(string(userPkg), string(pkgPrefix)) {
			return nil
		}
	}

	return VisibilityError{
		user:   user,
		entity: entity,
		rules:  r,
	}
}

type VisibilityErrors []error

func (errs VisibilityErrors) Error() string {
	bldr := &strings.Builder{}
	_, _ = fmt.Fprint(bldr, "visibility constraints have been violated:")

	for _, err := range errs {
		_, _ = fmt.Fprint(bldr, "\n - ", err)
	}

	return bldr.String()
}

type VisibilityError struct {
	user   protoreflect.Descriptor
	entity protoreflect.Descriptor
	rules  *rules
}

func (verr VisibilityError) Error() string {
	return fmt.Sprintf("%s: %s includes %s (%s)",
		verr.user.ParentFile().Path(),
		verr.user.FullName(),
		verr.entity.FullName(),
		verr.rules.mode,
	)
}
