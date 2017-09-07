# What 
Implementation of <https://medium.com/smsjunk/handling-1-million-requests-per-minute-with-golang-f70ac505fcaa>
# How to use
for x in {1..50}; do curl -X "PUT" --data "{\"command\" : \"test $x\"}" http://127.0.0.1:8080/stack; done
