# Intro
MULTIEX is a user-friendly reverse proxy which expose multi ports on a local machine passing through NAT.**Only support TCP currently.**
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
**1. Build executables inside 'cmd' folder(Prebuilt binary only supports linux)**

**2. Start MultiEx server at public host**

```bash
$ ./server -token a
```
**3. Start MultiEx client at local host behind NAT**

```bash
$ ./client -remotePort [server public ip]:[server port,default 8070] -portMap 2222-1800,2223-1100 -token a
```
**4. Access public port 2222/2223 to visit local port 1080/1100**

**5. Set open files limit(Optional)**

```bash
$ ulimit -n 65535
```

