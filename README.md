# Ignore Scanner

A small go utility to scan the directories for ignore patterns typically .dockerignore, .gitingore 
and list the files that are not ignored

# Build locally

```shell script
git clone https://github.com/kameshsampath/goignorescanner
# $PROJECT_HOME
cd goignorescanner
./hack/build.sh
```

# List files

```
./out/scanner -d pkg/scanner/testdata/starignore/
```

A sample output of the scanner, that list files which are not ignored by .dockerignore available as part of the folder `$PROJECT_HOME/pkg/scanner/testdata/starignore/`

```shell script
pkg/scanner/testdata/starignore/README.md
pkg/scanner/testdata/starignore/target/foo-runner.jar
pkg/scanner/testdata/starignore/target/lib/one.jar
pkg/scanner/testdata/starignore/target/quarkus-app/one.txt
```