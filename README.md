# sim
simple im

```text
netstat -n | awk '/^tcp/ {++State[$NF]} END {for(i in State) print i, State[i]}'
```