# Git LFS Server API compliance test utility

This package exists to provide automated testing of server API implementations, 
to ensure that they conform to the behaviour expected by the client. You can
run this utility against any server that implements the Git LFS API. 

## Automatic or data-driven testing

This utility is primarily intended to test the API implementation, but in order
to correctly test the responses, the tests have to know what objects exist on
the server already and which don't. 

In 'automatic' mode, the tests require that both the API and the content server
it links to via upload and download links are both available & free to use. 
The content server must be empty at the start of the tests, and the tests will
upload some data as part of the tests. Therefore obviously this cannot be a
production system.

Alternatively, in 'data-driven' mode, the tests must be provided with a list of 
object IDs that already exist on the server (minimum 10), and a list of other
object IDs that are known to not exist. The test will use these IDs to 
construct its data sets, will only call the API (not the content server), and
thus will not update any data - meaning you can in theory run this against a 
production system. 

## Calling the test tool

```
git-lfs-test-server-api [--url=<apiurl> | --clone=<cloneurl>] 
                        [<oid-exists-file> <oid-missing-file>]
                        [--save=<fileprefix>]
```

|Argument|Purpose|
|------|-------|
|`--url=<apiurl>`|URL of the server API to call. This must point directly at the API root and not the clone URL, and must be HTTP[S]. You must supply either this argument or the `--clone` argument|
|`--clone=<cloneurl>`|The clone URL from which to derive the API URL. If it is HTTP[S], the test will try to find the API at `<cloneurl>/info/lfs`; if it is an SSH URL, then the test will call git-lfs-authenticate on the server to derive the API (with auth token if needed) just like the git-lfs client does. You must supply either this argument or the `--url` argument|
|`<oid-exists-file> <oid-missing-file>`|Optional input files for data-driven mode (both must be supplied if this is used); each must be a file with `<oid> <size_in_bytes>` per line. The first file must be a list of oids that exist on the server, the second must be a list of oids known not to exist. If supplied, the tests will not call the content server or modify any data. If omitted, the test will generate its own list of oids and will modify the server (and expects that the server is empty of oids at the start)|
|`--save=<fileprefix>`|If specified and no input files were provided, saves generated test data in the files `<fileprefix>_exists` and `<fileprefix>_missing`. These can be used as parameters to subsequent runs if required, if the server content remains unchanged between runs.|
## Authentication

Authentication will behave just like the git-lfs client, so for HTTP[S] URLs the
git credential helper system will be used to obtain logins, and for SSH URLs,
keys can be used to automate login. Otherwise you will receive prompts on the
command line.

