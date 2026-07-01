# Model

## Metadata

- context: ctx-0001

## Definition

The set of all Entities (con-0001) loaded from a root directory, indexed by
Identifier (con-0002), together with the map from each identifier to the file it was
read from. The model is what every reporting capability reads. Loading (cap-0002)
builds the model and tolerates malformed content, so a partial model loads even when
some files have problems.

## States

The model has no lifecycle state. It is rebuilt from the files on each command.

## Relationships

The model holds every Entity (con-0001) and resolves References (con-0003) between
them by Identifier (con-0002). Every Graph (con-0005), Capability bundle (con-0006),
and Review packet (con-0007) is produced from a model.
