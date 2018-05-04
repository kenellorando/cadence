#!/bin/bash

generator/changelog-generator.sh -t .. -o content -i hashes

cat ./header ./content ./footer > ../public/changelog.html

rm -f ./content
