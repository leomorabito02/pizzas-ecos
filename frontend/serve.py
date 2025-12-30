#!/usr/bin/env python3
import http.server
import socketserver
import os

os.chdir(os.path.dirname(__file__))

PORT = 3000
Handler = http.server.SimpleHTTPRequestHandler

with socketserver.TCPServer(("", PORT), Handler) as httpd:
    print(f"Serving at http://localhost:{PORT}/")
    httpd.serve_forever()
