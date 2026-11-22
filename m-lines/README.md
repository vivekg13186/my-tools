# lines

A lightweight command-line utility written in Go for processing text
files one line at a time.

`lines` reads an input file, applies one or more actions in the order
supplied, and writes the result to standard output or an output file. It
uses only the Go standard library and requires no third-party
dependencies.

## Features

-   Sort lines in ascending or descending lexicographical order.
-   Remove duplicate lines.
-   Remove empty or whitespace-only lines.
-   Trim leading and trailing whitespace.
-   Add a prefix or suffix to every line.
-   Keep only lines matching a regular expression.
-   Apply an action conditionally using `-if REGEX`.
-   Chain multiple actions in a single command.
-   Write results to a file with `-o`.

## Requirements

-   Go 1.20 or later (the program uses only the standard library).
-   Windows, Linux, or macOS.

## Build

Save the source code as `main.go`, then build it:

``` sh
go build -o lines main.go
```

On Windows:

``` powershell
go build -o lines.exe main.go
```

You can also run it without building an executable:

``` sh
go run main.go input.txt trim rm-empty
```

## Basic syntax

``` text
lines INPUT_FILE [OPTIONS] [ACTIONS...]
```

-   `INPUT_FILE`: file to read.
-   `-o OUTPUT_FILE`: write the processed lines to this file. If
    omitted, output is written to standard output.
-   `ACTIONS`: one or more actions. Actions are executed from left to
    right.

The input file is not modified unless you explicitly choose it as the
output file. It is generally safer to write to a different output file.

## Actions

### `asc`

Sort all lines in ascending lexicographical order.

``` sh
lines names.txt asc
```

Example input:

``` text
Charlie
Alice
Bob
```

Output:

``` text
Alice
Bob
Charlie
```

### `dsc`

Sort all lines in descending lexicographical order.

``` sh
lines names.txt dsc
```

Output:

``` text
Charlie
Bob
Alice
```

Sorting is lexicographical (string) sorting, not numeric sorting. For
example, `10` sorts before `2` in ascending order.

### `dedup`

Remove duplicate lines. The first occurrence is kept; later identical
occurrences are removed.

``` sh
lines names.txt dedup
```

Example input:

``` text
apple
orange
apple
banana
orange
```

Output:

``` text
apple
orange
banana
```

Deduplication compares the complete line exactly. It is case-sensitive
and does not automatically trim whitespace. To remove duplicates after
trimming, chain the actions:

``` sh
lines names.txt trim dedup
```

### `rm-empty`

Remove empty lines and lines containing only whitespace.

``` sh
lines input.txt rm-empty
```

Example input:

``` text
first line


second line
```

Output:

``` text
first line
second line
```

### `trim`

Remove leading and trailing whitespace from each line. Internal
whitespace is preserved.

``` sh
lines input.txt trim
```

Example input:

``` text
   hello world
goodbye   
```

Output:

``` text
hello world
goodbye
```

A common combination is:

``` sh
lines input.txt trim rm-empty
```

This trims every line and then removes lines that became empty.

### `prefix TEXT`

Add `TEXT` to the beginning of every line.

``` sh
lines input.txt prefix "INFO: "
```

Example input:

``` text
Starting
Finished
```

Output:

``` text
INFO: Starting
INFO: Finished
```

Use quotes when the prefix contains spaces or shell-special characters.

### `suffix TEXT`

Add `TEXT` to the end of every line.

``` sh
lines input.txt suffix ";"
```

Example input:

``` text
apple
orange
```

Output:

``` text
apple;
orange;
```

### `filter REGEX`

Keep only lines that match the supplied regular expression. Lines that
do not match are discarded.

``` sh
lines application.log filter "ERROR"
```

Example input:

``` text
INFO Application started
ERROR Connection failed
DEBUG Retrying
ERROR Request timed out
```

Output:

``` text
ERROR Connection failed
ERROR Request timed out
```

More examples:

``` sh
# Lines beginning with ERROR
lines application.log filter "^ERROR"

# Lines containing a number
lines input.txt filter "[0-9]+"

# Lines ending in .go
lines files.txt filter "\.go$"

# Lines containing either ERROR or WARN
lines application.log filter "ERROR|WARN"
```

The regular expression syntax is Go's RE2 syntax. See [Go regexp
documentation](https://pkg.go.dev/regexp) for supported expressions.

## Conditional actions: `-if REGEX`

`-if REGEX` makes the **next action only** apply to lines that match the
condition.

Syntax:

``` text
lines INPUT_FILE -if CONDITION_REGEX ACTION [ACTION_ARGUMENT]
```

For example:

``` sh
lines application.log -if "^ERROR" prefix "[ALERT] "
```

Only lines beginning with `ERROR` receive the prefix. Other lines remain
unchanged.

Input:

``` text
INFO Application started
ERROR Connection failed
WARN Retrying
ERROR Request timed out
```

Output:

``` text
INFO Application started
[ALERT] ERROR Connection failed
WARN Retrying
[ALERT] ERROR Request timed out
```

### Conditional examples

``` sh
# Add a prefix only to warning lines
lines app.log -if "^WARN" prefix "[WARNING] "

# Add a suffix only to lines containing a request ID
lines app.log -if "request_id=" suffix " [checked]"

# Remove duplicates only among lines containing ERROR
lines app.log -if "ERROR" dedup

# Sort only matching lines; nonmatching lines stay in their original positions
lines app.log -if "^ERROR" asc
```

`-if` is consumed by the immediately following action. It does not
automatically apply to subsequent actions.

For example:

``` sh
lines app.log -if "^ERROR" prefix "[ALERT] " trim
```

Here, only `prefix` is conditional; `trim` applies to every line.

## Combining actions

Actions run from left to right. The output of one action becomes the
input to the next action.

``` sh
# Trim whitespace, remove blank lines, remove duplicates, then sort
lines input.txt trim rm-empty dedup asc

# Keep error lines, remove duplicates, and sort
lines app.log filter "^ERROR" dedup asc

# Add a prefix and suffix to each line
lines input.txt prefix "[" suffix "]"

# Keep warning lines, trim them, and sort
lines app.log filter "^WARN" trim asc

# Add a prefix to error lines, then remove empty lines
lines app.log -if "^ERROR" prefix "[ALERT] " rm-empty
```

## Writing output to a file

Use `-o` to save the result instead of printing it to the terminal.

``` sh
lines input.txt -o output.txt trim
```

You can place `-o` before or after the actions:

``` sh
lines input.txt -o output.txt trim dedup asc
lines input.txt trim dedup asc -o output.txt
```

If the output file already exists, it is overwritten.

## Practical recipes

### Clean a text file

``` sh
lines raw.txt -o clean.txt trim rm-empty dedup
```

Trims whitespace, removes blank lines, and removes duplicates.

### Extract errors from a log

``` sh
lines server.log -o errors.log filter "ERROR"
```

### Extract errors and warnings

``` sh
lines server.log -o important.log filter "ERROR|WARN"
```

### Add timestamps or labels

``` sh
lines tasks.txt -o labelled.txt prefix "[TODO] "
```

### Sort a list of email addresses

``` sh
lines emails.txt -o sorted-emails.txt trim rm-empty dedup asc
```

### Mark only failed requests

``` sh
lines requests.log -o marked.log -if "failed" suffix " [REVIEW]"
```

### Extract matching lines and clean them

``` sh
lines server.log -o errors.txt filter "ERROR" trim dedup asc
```

## Action reference

  Action       Argument   Behaviour
  ------------ ---------- ----------------------------------------------
  `asc`        None       Sort lines ascending
  `dsc`        None       Sort lines descending
  `dedup`      None       Remove duplicate lines, keeping the first
  `rm-empty`   None       Remove empty and whitespace-only lines
  `trim`       None       Trim leading and trailing whitespace
  `prefix`     `TEXT`     Add text to the start of each line
  `suffix`     `TEXT`     Add text to the end of each line
  `filter`     `REGEX`    Keep only lines matching the regex
  `-if`        `REGEX`    Apply the next action only to matching lines

## Notes and limitations

-   Actions are processed in the order given on the command line.
-   Matching and deduplication are case-sensitive.
-   `asc` and `dsc` use lexicographical string ordering.
-   `filter REGEX` removes nonmatching lines.
-   `-if REGEX` does not remove nonmatching lines; it leaves them
    unchanged for that action.
-   Regular expressions use Go's standard `regexp` package (RE2 syntax).
-   The line reader uses a 10 MiB maximum scanner token size. A single
    line longer than this limit will cause a read error.
-   Output is written one line at a time with a newline after each line.

## License

Choose and add a license before distributing this project. For example,
MIT.
