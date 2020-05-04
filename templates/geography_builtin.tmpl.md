Implement `{{ index .Args "function_name" }}` on arguments {{ index .Args "function_args" }}, which should follow the [PostGIS implementation](https://postgis.net/docs/{{ index .Args "function_name" }}.html).

For Geography builtins, please do the following:
* Add a relevant helper function in [`pkg/geo/geogfn`](https://github.com/cockroachdb/cockroach/tree/master/pkg/geo/geogfn). Add unit tests here - you can run through example test cases and make sure that PostGIS and CRDB return the same result within a degree of accuracy (1cm for geography).
* Create a new builtin that references this function in [`pkg/sql/sem/builtins/geo_builtins.go`](https://github.com/cockroachdb/cockroach/blob/master/pkg/sql/sem/builtins/geo_builtins.go).
* Modify the tests in [`pkg/sql/logictest/testdata/logic_test/geospatial`](https://github.com/cockroachdb/cockroach/blob/master/pkg/sql/logictest/testdata/logic_test/geospatial) to call this functionality at least once. You can call `make testbaselogic FILES='geospatial' TESTFLAGS='-rewrite'` to regenerate the output manually.
* Ensure the documentation is regenerated by calling `make buildshort`.
* Submit your PR - make sure to follow [creating your first PR](https://wiki.crdb.io/wiki/spaces/CRDB/pages/181633464/Your+first+CockroachDB+PR]).

{{if ne (index .Args "expected_work") "" }}
The following guidance has been issued on implementing this function:

{{index .Args "expected_work"}}
{{end}}

<sub>:robot: This issue was synced with by [gsheets-to-github-issues](https://github.com/cockroachlabs/gsheet-to-github-issues) by [{{ .GithubUser }}]({{ .GithubUserURL }}).</sub>