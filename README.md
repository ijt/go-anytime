# Go Anytime

[![Go Report Card](https://goreportcard.com/badge/github.com/ijt/go-anytime)](https://goreportcard.com/report/github.com/ijt/go-anytime)
[![codecov](https://codecov.io/gh/ijt/go-anytime/branch/main/graph/badge.svg?token=GA1TIU7OHK)](https://codecov.io/gh/ijt/go-anytime)
[![GoDoc](https://godoc.org/github.com/ijt/go-anytime?status.svg)](https://godoc.org/github.com/ijt/go-anytime)
![](https://img.shields.io/badge/license-MIT-blue.svg)
[![Awesome](https://cdn.rawgit.com/avelino/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#date-and-time)

Natural date time parsing for Go. This package was originally forked from
[github.com/tj/go-naturaldate](https://github.com/tj/go-naturaldate) but has diverged so much that it needed a new name
to avoid confusion. Here are the largest differences:

1. The `go-anytime` module is written in terms of the [github.com/ijt/goparsify](https://github.com/ijt/goparsify) parser combinator module, rather than the [github.com/pointlander/peg](https://github.com/pointlander/peg) parsing module. That made its development and debugging easier, and also means that its parsers can be use within other parsers that use `ijt/goparsify`.
2. Ranges can be parsed using `ParseRange` or `RangeParser`, for example `"from 3 feb 2022 until 6 oct 2022"`.
3. Dates/times and ranges can be replaced in strings using the funcs `ReplaceTimesByFunc`, `ReplaceRangesByFunc`, and `ReplaceDateRangesByFunc`.
4. Strings can be partitioned into time and non-time parts using the funcs `PartitionTimes` and `PartitionTimesByFuncs`.

## Examples

Here are some examples of expressions that can be parsed by `anytime.Parse()` or `anytime.Parser`:

- now
- today
- yesterday
- 5 minutes ago
- three days ago
- last month
- next month
- one year from now
- yesterday at 10am
- last sunday at 5:30pm
- next sunday at 22:45
- next January
- last February
- next December 25th at 7:30am
- next December 25th at 7:30am UTC-7
- November 3rd, 1986 at 4:30pm
- january 2017
- january, 2017
- oct 7, 1970
- oct 7 1970
- 7 oct 1970
- 7 oct, 1970
- September 17, 2012 UTC+7
- September 17, 2012
- 03 February 2013
- 2 July 2013
- 2014/3/31
- 2014/3/31 UTC
- 2014/3/31 UTC+1
- 2014/03/31
- 2014/03/31 UTC-1
- 2014-04-26
- 2014-4-26
- 2014-4-6
- 31/3/2014 UTC-8
- 31-3-2014 UTC-8
- 31/3/2014
- 31-3-2014
- January
- december 20
- thursday at 23:59
- See the [tests](./anytime_test.go) for more examples

## Range examples

Here are some examples of expressions that can be parsed by `anytime.ParseRange()` or `anytime.RangeParser`:

- from 3 feb 2022 to 6 oct 2022
- 3 feb 2022 to 6 oct 2022
- from 3 feb 2022 until 6 oct 2022
- from tuesday at 5pm -12:00 until thursday 23:52 +14:00
- last year
- today
- next week
