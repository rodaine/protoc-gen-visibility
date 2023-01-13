package runner

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (r *runner) checkFile(file *protogen.File) (errs []error) {
	for _, svc := range file.Services {
		errs = append(errs, r.checkService(svc)...)
	}

	for _, msg := range file.Messages {
		errs = append(errs, r.checkMessage(msg)...)
	}

	return errs
}

func (r *runner) checkMessage(msg *protogen.Message) (errs []error) {
	for _, fld := range msg.Fields {
		switch fld.Desc.Kind() {
		case protoreflect.MessageKind:
			errs = append(errs, r.checkEntity(fld.Desc, fld.Message.Desc)...)
		case protoreflect.EnumKind:
			errs = append(errs, r.checkEntity(fld.Desc, fld.Enum.Desc)...)
		}
	}

	for _, m := range msg.Messages {
		errs = append(errs, r.checkMessage(m)...)
	}

	return errs
}

func (r *runner) checkService(svc *protogen.Service) (errs []error) {
	for _, mtd := range svc.Methods {
		errs = append(errs, r.checkEntity(mtd.Desc, mtd.Input.Desc)...)
		errs = append(errs, r.checkEntity(mtd.Desc, mtd.Output.Desc)...)
	}

	return errs
}

func (r *runner) checkEntity(user protoreflect.Descriptor, entity protoreflect.Descriptor) (errs []error) {
	if err := r.lookup[entity.FullName()].checkUsage(user, entity); err != nil {
		errs = append(errs, err)
	}

	return errs
}
