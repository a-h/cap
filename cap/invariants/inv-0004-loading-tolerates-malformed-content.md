# Loading tolerates malformed content

## Metadata

- capabilities:

## Description

Loading (cap-0002) builds as much of the Model (con-0004) as the files allow and
records each problem it finds, rather than failing on the first malformed file. A
missing section, an unparseable identifier, or a dangling reference degrades to a
recorded problem, and the entities that did parse are still available to every report.

## Capabilities

- cap-0002

## Rationale

An author edits the model incrementally and runs reports against work in progress. If
loading stopped at the first problem, a single unfinished file would blind every
report to the rest of the model. Tolerating malformed content lets validation list
every problem at once and lets other reports work on the parts that are sound.
