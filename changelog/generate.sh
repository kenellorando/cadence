#!/bin/bash

generator/changelog-generator.sh -t .. -o content -i hashes

cat ./header ./content ./footer > ../public_html/changelog.html

rm -f ./content
