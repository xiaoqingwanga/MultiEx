# Intro
MULTIEX is a reverse proxy which expose multi ports on a local machine passing through NAT.
# Usage
```bash
$ ./client -h                          
Usage of ./client:
  -logLevel string
    	the log level of this program. (default "INFO")
  -logTo string
    	the location where logs save. Empty value and stdout have special meaning (default "stdout")
  -portMap string
    	Port map represent mapping between host. e.g. '2222-22' represents expose local port 22 at public port 2222. Multi mapping split by comma. (default "2222-22")
  -remotePort string
    	the public server ip:port listening for MultiEx client.
  -token string
    	Token is the credential client should hold to connect server.Server doesn't have token default.
$ ./server -h    
  Usage of ./server:
    -clientPort string
      	the port listening for MultiEx client. (default ":8070")
    -logLevel string
      	the log level of this program. (default "INFO")
    -logTo string
      	the location where logs save. Empty value and stdout have special meaning (default "stdout")
    -token string
      	Token is the credential client should hold to connect this server.Server doesn't have token default.

```
**1. build executable inside 'cmd' folder**

**2. start MultiEx server at public host**
```bash
$ ./server -token a
```
**3. start MultiEx client at local host behind NAT**
```bash
$ ./client -remotePort [ip]:[port] -portMap 2222-1800,2223-1100 -token a
```
**4. access public port 2222 to visit local port 22**