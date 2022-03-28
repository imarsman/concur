#!/bin/bash

term="$((1 + $RANDOM % 20))"

i=0
j=1
for (( c=0; c<term; c++ )) ; do
    (( n=i+j ))
    (( i=j ))
    (( j=n ))
#    echo $i
done

echo "${i}"
