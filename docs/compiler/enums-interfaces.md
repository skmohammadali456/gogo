# Step 14: enums and interfaces

`create enum State { Ready, Loading as String }` declares a session-scoped canonical enum. `State.Ready` is a canonical variant expression and a payload variant is constructed as `State.Loading("please wait")`. A variant is assignable to its parent enum, while enum and variant identity are nominal and deterministic. Enum types are available in annotations and type aliases. The private `types.Value` boundary validates enum payloads defensively.

`create interface User extends Identified { readonly id as Number, name? as String }` declares a frontend-friendly structural contract. Interfaces resolve to the existing canonical object type model, so object implementation checking, aliases, nested properties, unions, intersections, `Optional`, and `Result` use the established assignability rules. Required fields must exist, optional source fields cannot satisfy required target fields, and readonly source properties cannot satisfy mutable targets.

Multiple parents are accepted. Resolution is deterministic and diagnoses duplicate declarations, unresolved parents, cycles, and incompatible inherited/overridden properties. Enum equality checks and existing interface object discriminants use the Step 13 narrowing engine; refinements are branch-local and assignment resets a narrowed binding to its declared type.

The grammar is data-driven: English uses `enum`, `interface`, `extends`; Bengali uses `এনাম`, `ইন্টারফেস`, `প্রসারিত`; Hindi uses `एनम`, `इंटरफ़েস`, `विस्तार`. Diagnostics G3100–G3104 have English, Bengali, and Hindi catalog text.

There is no interpreter, exhaustive pattern matching, classes, or generics in this step.
