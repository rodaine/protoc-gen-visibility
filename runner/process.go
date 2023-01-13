package runner

import (
	"github.com/rodaine/protoc-gen-visibility/visibility/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

func (r *runner) processFile(file *protogen.File) {
	rset := r.processEntity(nil, file.Desc, visibility.E_FileRules)

	for _, enum := range file.Enums {
		r.processEnum(rset, enum)
	}

	for _, msg := range file.Messages {
		r.processMessage(rset, msg)
	}
}

func (r *runner) processEnum(parent *rules, enum *protogen.Enum) {
	r.processEntity(parent, enum.Desc, visibility.E_EnumRules)
}

func (r *runner) processMessage(parent *rules, msg *protogen.Message) {
	rset := r.processEntity(parent, msg.Desc, visibility.E_MessageRules)

	for _, enum := range msg.Enums {
		r.processEnum(rset, enum)
	}

	for _, m := range msg.Messages {
		r.processMessage(rset, m)
	}
}

func (r *runner) processEntity(parent *rules, desc protoreflect.Descriptor, ext *protoimpl.ExtensionInfo) *rules {
	spec, _ := proto.GetExtension(desc.Options(), ext).(*visibility.Specification)

	if spec == nil && parent != nil {
		r.lookup[desc.FullName()] = parent
		return parent
	}

	rset := processRules(desc, spec)
	r.lookup[desc.FullName()] = rset

	return rset
}
