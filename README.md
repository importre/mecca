mecca
=====

Collect projects on Github and Categorize.

Currently **mecca** gathers `Polymer` and `Android` repositories.


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

Polymer repositories
--------------------

### Polymer Repository
```
polymer fork:true stars:>=<The number of stars>
```

### Polymer Project
```
"platform.js" extension:html language:html repo:<Repository Name>
```


Android repositories
--------------------

### Android Repository
```
android fork:true stars:>=<The number of stars>
```

### Android Project
```
androidmanifest.xml in:path repo:<Repository Name>
```

### Android Library
```
"android-library" extension:gradle language:groovy repo:<Repository Name>
"android.library=true" language:ini repo:<Repository Name>
```

If both are not matched, check repo's description using regex `(?i)\blibrary\b` whether it has `library` literal or not.

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
