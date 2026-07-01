# A reference names an entity that exists

## Metadata

- capabilities:

## Description

Every Reference (con-0003) resolves to a loaded Entity (con-0001) of the kind its
identifier's prefix implies. A reference to a missing identifier, or to an identifier
whose kind does not match the section it appears in, is reported as a problem.

## Capabilities

- cap-0003

## Rationale

References are the edges of the Model (con-0004). A dangling or mistyped reference
breaks traceability silently: a report would omit a link that the author believed was
present. Surfacing the problem keeps the graph an author navigates faithful to what
they wrote.
