#!/bin/bash

for i in `seq 1 100`; do 
	echo "test round $i"
	rm -f *.cpnt; 
	go test -race -run 3A; 
	#go test -race -run TestOpenBD3A; 
	#go test -race -run TestCreateBD3A; 
	if [[ $? != "0" ]] ; then break; fi
done

