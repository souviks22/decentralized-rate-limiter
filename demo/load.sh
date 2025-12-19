#!/bin/bash
> targets.txt  # clear the file

for i in {1..1000}; do
  echo "GET http://localhost/api/ping" >> targets.txt
  echo "Ping-User-Id: $i" >> targets.txt
  echo "" >> targets.txt
done
