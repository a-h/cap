# A many-to-many link is declared on both entities

## Metadata

- capabilities:

## Description

Where two Entities (con-0001) share a many-to-many relationship, the reference
(con-0003) is declared on both. A capability names its invariant, and the invariant
names the capability back; the same holds for a capability and its specification, and
for a capability and its scenario. When only one side names the other, the link is
reported as asymmetric.

## Capabilities

- cap-0003

## Rationale

A link declared on one side only is usually a half-finished edit: an entity was
renamed or removed and its counterpart was not updated. Requiring both sides to
declare the link lets either file be read on its own and shows the whole relationship,
and lets validation catch the forgotten half.
