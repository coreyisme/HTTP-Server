# HTTP-Server
A HTTP clone in Golang
# Supported features
- Multiple HTTP compression schema (gzip, br, deflate,...)
- Adding _Keep-Alive_ tag to hold a connection for multiple **CONCURRENT reqs**
- Adding _Connection_ tag to be able to handle 1 or multiple **PERSISTENT connections**
- Shutdown gracefully
# Future extensions:
- HTTPS
- E-Tag caching
- HTTP/2
- Range requests
- WebSockets
