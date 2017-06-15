# Laurence Morgan .co.uk

This uses the Level 10 Fireball CMS for forum engine I had written quite
a number of years back. It was originally written in Perl but this is
the Go (Golang) port. However the code is still old and not the greatest
example of my coding; while the site itself works fine the code breaks a
number of Go idioms, not the easiest to read, and frankly is a bit of an
embarrassment. However throwing it on a public repo means I can simplify
my CI pipeline. So here it is. :)

# Copyright

There is intentionally no LICENCE file because I currently don't deem
this project to be good enough to open source (ie there are a plethora
of other CMS's out there). However if the interest is there then I am
open to the idea of changing the licencing to something more permissive.

# Build

docker build -t laurencemorgan .

# Import

docker pull lmorg/laurencemorgan:latest && docker run --publish 80:8080 \
  -v ~/uploads:/uploads --name laurencemorgan --rm lmorg/laurencemorgan:latest

# Run

docker run --publish 80:8080 -v /uploads:/uploads --name laurencemorgan \
  -e DB_USERNAME="$DB_USERNAME" -e DB_PASSWORD="$DB_PASSWORD" \
  --rm lmorg/laurencemorgan:latest
