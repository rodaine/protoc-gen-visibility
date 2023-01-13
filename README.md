# protoc-gen-visibility

_Set cascading visibility rules for Protocol Buffer files, messages, and enums._

## Example

```protobuf
// ./protos/widgets/v1/widget.proto
syntax = "proto3";
package widgets.v1;

import "visibility/v1/visibility.proto";

option (visibility.v1.file_rules) = {
  mode: MODE_PACKAGE, // entities in this file are package-private...
  packages: [         // ...and also visible to these packages:  
    "gizmos.v1", 
    "gadgets.*"
  ]
};

// ...
```

`protoc-gen-visibility` is executed like any other protoc plugin, but does not generate output files. The plugin errors if there is a violation of the visibility specificiation:

```shell
$ protoc --visibility_out=. ./protos/gadgets/v1/gadget.proto
# exits 0

$ protoc --visibility_out=. ./protos/gizmos/v2/gizmo.proto
--visibility_out: visibility constraints have been violated:
 - protos/gizmos/v2/gizmo.proto: gizmos.v2.Gizmo includes widgets.v1.Widget (MODE_PACKAGE)
# exits 1
```