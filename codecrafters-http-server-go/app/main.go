/*package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strings"












	"github.com/codecrafters-io/http-server-starter-go/internal/config"
	"github.com/codecrafters-io/http-server-starter-go/internal/constants"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit
var fileStorageDir string

func main() {
	s := Server{}
	s.Start()

	for {
		conn := s.Accept()

		// Each request are made by 3 parts SEPERATED BY CRLF:
		// Request line (status line)
		// Header
		// Request body (optional)
		if conn == nil {
			continue
		}
		go s.handleRequest(conn)
	}
}

type Server struct {
	listener net.Listener
	// conn     net.Conn
}

func (s *Server) handleRequest(conn net.Conn) {
	defer conn.Close()

	//In HTTP/1.1, the server could only hanlde 100 PERSISTENT connections at the same time with Keep-Alive timeout = 5s
	reqCount := 0

	for {
		reqCount++
		if reqCount > config.MaxConnection {
			fmt.Println("Max request reached, closing connection now...")
			break
		}

		buffer := make([]byte, 8192)
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Failed to handle client's request due: ", err.Error())
		}

		req := string(buffer[:n])

		lines := strings.Split(req, "\r\n")
		//HTTP request components
		method := strings.Split(lines[0], " ")[0]
		URL := strings.Split(lines[0], " ")[1]

		headers := make(map[string]string)
		for i := 1; i < len(lines); i++ {
			line := lines[i]

			// An empty line signifies the end of headers
			if line == "" {
				break
			}

			parts := strings.SplitN(line, ":", 2) // Split only on the first colon to handle values with colons
			if len(parts) == 2 {
				headerName := strings.TrimSpace(parts[0])
				headerValue := strings.TrimSpace(parts[1])
				headers[headerName] = headerValue
			} else {
				// Handle malformed header line if necessary
				fmt.Printf("Warning: Malformed header line: %s\n", line)
			}
		}
	}

	//HTTP response components
	reqBody := strings.TrimSuffix(lines[len(lines)-1], "\x00")

	var res string
	if URL == "/" {
		res = "HTTP/1.1 200 OK\r\n\r\n"
		conn.Write([]byte(res))
	} else if strings.HasPrefix(URL, "/echo") {
		path := strings.Split(URL, "/")
		sub_path := path[len(path)-1]
		sub_length := len(sub_path)
		var encode_header string //Used for "Accept-Encoding" header
		for _, value := range constants.CompressionTypes {
			if strings.Contains(headers["Accept-Encoding"], value) {
				encode_header = value
			}
		}
		// URL = "/echo/abcxyz"
		// path = ["echo", "abcxyz"]
		// sub_path = path[1] = "abcxyz"
		// sub_length = len(sub_path)[1] = "abcxyz"
		if encode_header == "" {
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", sub_length, sub_path)
			conn.Write([]byte(res))
		} else {
			var buf bytes.Buffer
			gzipWrite := gzip.NewWriter(&buf)
			gzipWrite.Write([]byte(sub_path))
			gzipWrite.Close()
			compressed_message := buf.String()
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: %s\r\nContent-Length: %d\r\n\r\n%v", encode_header, len(compressed_message), compressed_message)
			conn.Write([]byte(res))
		}
	} else if URL == "/user-agent" {
		header := lines[2]
		path := strings.Split(header, " ")
		sub_path := path[len(path)-1]
		sub_length := len(sub_path)
		// header = "User-Agent: foobar/1.2.3"
		// path = ["User-Agent:", "foobar/1.2.3"]
		// sub_path = "foobar/1.2.3"
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", sub_length, sub_path)
		conn.Write([]byte(res))
	} else if URL[:6] == "/files" {
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "--directory" && i+1 < len(os.Args) {
				fileStorageDir = os.Args[i+1]
				break
			}
		}
		if method == "GET" {
			file_path := fmt.Sprintf("%s/%s", fileStorageDir, strings.TrimPrefix(URL, "/files/"))
			file, err := os.ReadFile(file_path)
			if err != nil {
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
				conn.Write([]byte(res))
			} else {
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file), string(file))
				conn.Write([]byte(res))
			}
			// filepath = "/files/{filename}"
			// filename = ["files", "{filename}"]
			// file = "{filename}"
		} else if method == "POST" {
			file_path := fmt.Sprintf("%s/%s", fileStorageDir, strings.TrimPrefix(URL, "/files/"))
			err := os.WriteFile(file_path, []byte(reqBody), 0644)
			if err != nil {
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
				conn.Write([]byte(res))
			} else {
				res = "HTTP/1.1 201 Created\r\n\r\n"
				conn.Write([]byte(res))
			}
		}

	} else {
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
		conn.Write([]byte(res))
	}

	fmt.Println(res)
}

func (s *Server) Start() {
	s.Listen()
	// s.conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}

func (s *Server) Listen() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind port 4221")
		os.Exit(1)
	}

	s.listener = l
}

func (s *Server) Accept() net.Conn {
	conn, err := s.listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		return nil
	}
	return conn
}

func (s *Server) Close() {
	err := s.listener.Close()
	if err != nil {
		fmt.Println("Error closing connection: ", err.Error())
		os.Exit(1)
	}
}
*/

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/internal/config"
	"github.com/codecrafters-io/http-server-starter-go/internal/constants"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

var _ = net.Listen
var _ = os.Exit
var fileStorageDir string

func main() {
	s := Server{}
	s.Start()

	for {
		conn := s.Accept()

		if conn == nil {
			continue
		}
		// Each request are made by 3 parts SEPERATED BY CRLF:
		// Request line (status line)
		// Header
		// Request body (optional)
		go s.handleRequest(conn)
	}
}

type Server struct {
	listener net.Listener
}

func (s *Server) handleRequest(conn net.Conn) {
	defer conn.Close()

	//In HTTP/1.1, the server can only handle ~100 PERSISTENT connections with keep - alive timout = 5s
	requestCount := 0

	for {
		requestCount++
		if requestCount > config.MaxConnection {
			fmt.Println("Max requests reached, closing connection.")
			break
		}

		conn.SetReadDeadline(time.Now().Add(config.KeepAliveTimeout))

		buffer := make([]byte, 8192)
		n, err := conn.Read(buffer)

		// Handle read errors or client closing connection
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Connection timed out, closing.")
			} else if err == io.EOF {
				fmt.Println("Client closed connection gracefully.")
			} else {
				fmt.Println("Failed to read client's request:", err.Error())
			}
			break
		}

		req := string(buffer[:n])

		lines := strings.Split(req, "\r\n")
		if len(lines) < 1 || strings.TrimSpace(lines[0]) == "" {
			fmt.Println("Received empty or malformed request line, closing connection.")
			break
		}

		// Parse Request Line
		requestLineParts := strings.Split(lines[0], " ")
		if len(requestLineParts) < 2 {
			fmt.Println("Malformed request line:", lines[0])
			sendResponse(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
			break
		}
		method := requestLineParts[0]
		URL := requestLineParts[1]

		// Parse Headers
		headers := make(map[string]string)
		bodyStartIndex := -1
		for i := 1; i < len(lines); i++ {
			line := lines[i]

			if line == "" {
				bodyStartIndex = i + 1
				break
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headerName := strings.TrimSpace(parts[0])
				headerValue := strings.TrimSpace(parts[1])
				headers[headerName] = headerValue
			} else {
				fmt.Printf("Warning: Malformed header line: %s\n", line)
			}
		}

		// Extract Request Body
		reqBody := ""
		if bodyStartIndex != -1 && bodyStartIndex < len(lines) {
			reqBody = strings.Join(lines[bodyStartIndex:], "\r\n")
			reqBody = strings.TrimSuffix(reqBody, "\x00")
		}

		connectionHeaderValue := "close"

		requestedConnection := strings.ToLower(headers["Connection"])

		if requestedConnection == "keep-alive" || (requestedConnection == "" && strings.HasSuffix(lines[0], "HTTP/1.1")) {
			connectionHeaderValue = "keep-alive"
		}
		if requestedConnection == "close" {
			connectionHeaderValue = "close"
		}

		var resHeaders []string
		resHeaders = append(resHeaders, fmt.Sprintf("Connection: %s", connectionHeaderValue))

		if URL == "/" {
			sendResponse(conn, "HTTP/1.1 200 OK\r\n"+strings.Join(resHeaders, "\r\n")+"\r\n\r\n")
		} else if strings.HasPrefix(URL, "/echo") {
			pathParts := strings.Split(URL, "/")
			subPath := pathParts[len(pathParts)-1] // "abcxyz" <- "/echo/abcxyz"

			var encodeHeader string

			acceptEncoding := headers["Accept-Encoding"]
			if acceptEncoding != "" {
				for _, value := range constants.CompressionTypes {
					if strings.Contains(acceptEncoding, value) {
						encodeHeader = value
						break
					}
				}
			}

			var responseBody []byte
			var contentEncodingHeader string
			var contentLength int

			if encodeHeader == "" {
				responseBody = []byte(subPath)
				contentLength = len(responseBody)
			} else {
				var buf bytes.Buffer
				gzipWriter := gzip.NewWriter(&buf) //
				_, err := gzipWriter.Write([]byte(subPath))
				if err != nil {
					fmt.Println("Error compressing data:", err)
					sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
					return
				}
				gzipWriter.Close()
				responseBody = buf.Bytes()
				contentEncodingHeader = fmt.Sprintf("Content-Encoding: %s", encodeHeader)
				contentLength = len(responseBody)
				resHeaders = append(resHeaders, contentEncodingHeader)
			}

			resHeaders = append(resHeaders, "Content-Type: text/plain")
			resHeaders = append(resHeaders, fmt.Sprintf("Content-Length: %d", contentLength))

			response := fmt.Sprintf("HTTP/1.1 200 OK\r\n%s\r\n\r\n%s", strings.Join(resHeaders, "\r\n"), responseBody)
			conn.Write([]byte(response))

		} else if URL == "/user-agent" {
			userAgent := headers["User-Agent"]
			if userAgent == "" {
				userAgent = "N/A"
			}
			contentLength := len(userAgent)
			resHeaders = append(resHeaders, "Content-Type: text/plain")
			resHeaders = append(resHeaders, fmt.Sprintf("Content-Length: %d", contentLength))
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\n%s\r\n\r\n%s", strings.Join(resHeaders, "\r\n"), userAgent)
			conn.Write([]byte(response))

		} else if strings.HasPrefix(URL, "/files/") {
			currentFileStorageDir := ""
			for i := 1; i < len(os.Args); i++ {
				if os.Args[i] == "--directory" && i+1 < len(os.Args) {
					currentFileStorageDir = os.Args[i+1]
					break
				}
			}
			if currentFileStorageDir == "" {
				fmt.Println("Error: --directory argument not provided.")
				sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
				return
			}

			filename := strings.TrimPrefix(URL, "/files/")
			filePath := fmt.Sprintf("%s/%s", currentFileStorageDir, filename)

			if method == "GET" {
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					sendResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				} else {
					resHeaders = append(resHeaders, "Content-Type: application/octet-stream")
					resHeaders = append(resHeaders, fmt.Sprintf("Content-Length: %d", len(fileContent)))
					response := fmt.Sprintf("HTTP/1.1 200 OK\r\n%s\r\n\r\n%s", strings.Join(resHeaders, "\r\n"), fileContent)
					conn.Write([]byte(response))
				}
			} else if method == "POST" {
				err := os.WriteFile(filePath, []byte(reqBody), 0644)
				if err != nil {
					fmt.Println("Error writing file:", err)
					sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
				} else {
					sendResponse(conn, "HTTP/1.1 201 Created\r\n\r\n")
				}
			} else {
				sendResponse(conn, "HTTP/1.1 405 Method Not Allowed\r\n\r\n")
			}

		} else {
			sendResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		}

		// If keep-alive was not requested or max requests reached, break the loop
		if requestedConnection == "close" || requestCount >= config.MaxConnection {
			fmt.Println("Connection will close (either not requested or max requests reached).")
			break
		}
		fmt.Printf("Keeping connection alive for next request (%d/%d).\n", requestCount, config.MaxConnection)
	}
}

// Helper to send a simple response
func sendResponse(conn net.Conn, response string) {
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response:", err)
	}
}

func (s *Server) Start() {
	s.Listen()
}

func (s *Server) Listen() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind port 4221")
		os.Exit(1)
	}
	s.listener = l
}

func (s *Server) Accept() net.Conn {
	conn, err := s.listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		return nil
	}
	return conn
}

func (s *Server) Close() {
	err := s.listener.Close()
	if err != nil {
		fmt.Println("Error closing connection: ", err.Error())
		os.Exit(1)
	}
}
