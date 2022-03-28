#!/bin/bash

# e.g. time concur './fibonacci.sh' -a '{1..1000}' \
#        -A '{printf "%-10s", "cycles:"$1; printf "%-10s ",  " result:"$2}'

term="$((1 + $RANDOM % 30))"

i=0
j=1
for (( c=0; c<term; c++ )) ; do
    (( n=i+j ))
    (( i=j ))
    (( j=n ))
done

echo "${term} ${i}"
