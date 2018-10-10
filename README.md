# backend-finder
Find real backend server by scanning potential CIDR lists

```go run backend_finder.go <domain> <search> <file_CIDRs>```

  where:
  * domain = domain that served by backend server
  * search = unique string pattern on the served website
  * file_CIDRs = file with CIDRs one per line`
  

##### TODO
- [x] async channels
- [x] pattern scan
- [x] IP list based on CIDRs from file
- [x] timeouts & delays
- [ ] TLS/https support
