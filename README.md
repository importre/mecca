mecca
=====

Collect projects on Github and Categorize.

Currently **mecca** gathers Android repositories only.


Usage
-----

Set your github access token to `GITHUB_TOKEN` in `config.go`.  
Run commands as follows.

```sh
$ go build   # run `go get` if you don't have 3rd packages.
$ ./mecca -h # see help message.
$ ./mecca
```

The results(JSON files) will be generated in `data` directory.


Android repositories
--------------------

### Category and Rules (Queries)

#### All
```
android fork:true stars:>=<The number of stars>
```

#### Android Library
```
"android-library" extension:gradle language:groovy repo:<Repository Name>
"android.library=true" language:ini repo:<Repository Name>
```

#### Android-L
```
"compileSdkVersion android-l" extension:gradle language:groovy repo:<Repository Name>
```

#### Wearable
```
"com.google.android.support:wearable" extension:gradle language:groovy repo:<Repository Name>
```


Future works
------------

- Service http://importre.github.io/mecca
- New categories ?


References
----------

- [Searching Github][Searching Github]


License
-------

See `LICENSE` file.

[Searching Github]: https://help.github.com/articles/searching-github
