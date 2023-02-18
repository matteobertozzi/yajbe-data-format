#!/bin/bash

VERSION=0.9.0-SNAPSHOT
mvn deploy:deploy-file -e \
    -Dfile=./target/jackson-dataformat-yajbe-${VERSION}.jar \
    -DpomFile=./pom.xml \
    -DrepositoryId=github \
    -Durl=https://maven.pkg.github.com/matteobertozzi/yajbe-data-format \
    -Dtoken=GH_TOKEN
