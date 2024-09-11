# GopherDNS
An attempt to build a custom DNS server in Golang for learning purpose.


We need to add domains and thier ipv4 addresses in the ```domains.yaml``` file located at root of project.

Run the code as follows :
```go run main.go --dnsPort 1053 --serverRefreshTime 5 --ttlForResponse 5```


The server keeps checking for updated IPs in the file periodically , after every x seconds as set by the ```serverRefreshTime``` flag. Hence, it remains updated with IPs set.

Future task is to maintain a TTL based cache and look for changes in IPs from file and update the cache.

Following is an example run :
![command](https://raw.githubusercontent.com/hkiiita/GopherDNS/main/docs/screenshots/command.png)

Following is a response received by DIG command.
![screenshot](https://raw.githubusercontent.com/hkiiita/GopherDNS/main/docs/screenshots/screenshot1.png)






