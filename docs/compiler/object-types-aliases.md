# Step 11: object types and aliases

Step 11 keeps `internal/types.Type` as GOGO's only type representation. `Object`
is the Step 11 name for the existing canonical record representation: it has
string-named fields sorted deterministically, plus optional and readonly flags.
`Object{readonly name: String, tag?: String}` is valid in an annotation.

Object equality compares the full sorted field set: field name, canonical field
type, optionality, readonly status, and (when present) the index signature's
key type, value type, and readonly status. Source declaration order is not
semantic. Empty and duplicate field names are rejected.

Assignability is structural and distinct from equality. A source must supply
every required target property; an optional source property cannot satisfy a
required target property. Extra source properties are allowed. Property types
must be assignable recursively. A mutable source property can be viewed through
a readonly target property, but a readonly source cannot satisfy a mutable
target property. This prevents writes through a mutable view.

The canonical model supports string index signatures through `RecordWithIndex`.
A target index signature requires a compatible source signature, and every
explicit source property must satisfy its value type. Step 11 deliberately does
not add an arbitrary source spelling for index signatures because the existing
surface grammar has no settled bracketed type-member syntax; this is a parser
boundary, not a second type system.

`create type Name as Annotation` binds an alias only for the compilation
session. Aliases resolve directly to their target canonical type, including
other aliases, primitive, collection, and object annotations. Duplicate,
unresolved, and cyclic aliases report G3004/G3001 diagnostics rather than
creating a nominal type or hanging. `type` and `readonly` are data-driven
vocabulary entries: English (`type`, `readonly`), Bengali (`ধরন`, `শুধু_পঠন`),
and Hindi (`प्रकार`, `केवल_पढ़ने`) parse into the same AST and canonical types.

`types.RecordValue` remains the limited runtime boundary. It checks required
fields, allows absent optional fields, rejects unknown fields for closed
objects, and validates index-signature values when a canonical indexed object
is constructed. It is not an interpreter or object backend.
