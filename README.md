# EchoTool - Echo client and server

Heavily inspired by [EchoTool](https://github.com/PavelBansky/EchoTool).

Command line echo server and client. This tool is designed according to [RFC 862 specification]([http://www.ietf.org/rfc/rfc0862.txt?number=862]) for Echo protocol. It can operate as an echo server that sends back every incoming data. In client mode, it sends data to the server and checks whether they came back. This is a useful debugging tool for application development or network throughput checks.

- Server mode
- Client mode
- TCP and UDP protocol support
- Selectable destination and source port
- Selectable timeout
- Selectable echo pattern
- Just one file

## Usage and Options

```
EchoTool

Usage:
   echotool [options] [destination]

Options:
   -p protocol   TCP or UDP protocol
   -r port       Remote port number
   -o host       Local host for client/server
   -l port       Local port number for client/server (default: 0)
   -s            Server mode enabled (default: false)
   -c count      Number of echo requests to send (0 = infinite) (default: 5)
   -a pattern    Pattern to be sent for echo (default: DESKTOP-NGC8Q91)
   -t timeout    Connection timeout (default: 10s)
   -w deadline   Read/Write deadline (default: 5s)
   -i interval   Time interval between sending each echo request (default: 100ms)
   -d            Print various debugging information (default: false)
   -v            Print program version and exit
   -h            Print this help text and exit
```

## Examples

```
# For server mode listening on UDP port 1234 run following command:
echotool -s -p udp -l 1234

# On client machine run this:
echotool -p udp -r 1234 server.to-test.com

# You can specify outgoing local port by -l switch:
echotool -p udp -r 1234 -l 5678 server.to-test.com

# Number of attempts and timeouts can be set by -c and -t switch:
echotool -p udp -r 1234 -l 5678 -c 100 -t 10s server.to-test.com

# Use your own echo pattern with -a switch:
echotool -p udp -r 1234 -a Hello server.to-test.com
```

## Issues

Submit the [issues](https://github.com/attilabuti/echotool/issues) if you find any bug or have any suggestion.

## Contribution

Fork the [repo](https://github.com/attilabuti/echotool) and submit pull requests.

## License

This project is licensed under the [MIT License](https://github.com/attilabuti/echotool/blob/main/LICENSE).