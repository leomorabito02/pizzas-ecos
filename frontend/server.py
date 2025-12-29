#!/usr/bin/env python3
"""
Servidor simple para desarrollar el frontend
Ejecuta: python server.py
Luego abre: http://localhost:5000 o http://<TU_IP>:5000 desde el teléfono
"""

from http.server import HTTPServer, SimpleHTTPRequestHandler
import os
import sys
import socket

class MyHTTPRequestHandler(SimpleHTTPRequestHandler):
    def end_headers(self):
        # Evitar caché para desarrollo
        self.send_header('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
        super().end_headers()
    
    def log_message(self, format, *args):
        # Log más legible
        print(f"[{self.log_date_time_string()}] {format % args}")

if __name__ == "__main__":
    # Cambia al directorio frontend
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    
    port = 5000
    server_address = ('0.0.0.0', port)  # Escuchar en todas las interfaces
    httpd = HTTPServer(server_address, MyHTTPRequestHandler)
    
    # Obtener IP local
    try:
        hostname = socket.gethostname()
        local_ip = socket.gethostbyname(hostname)
    except:
        local_ip = "192.168.x.x"
    
    print(f"🚀 Frontend sirviendo en:")
    print(f"   - Local: http://localhost:{port}")
    print(f"   - Red: http://{local_ip}:{port} (desde otro dispositivo)")
    print(f"📁 Directorio: {os.getcwd()}")
    print(f"⚠️  Asegúrate que el backend está corriendo en http://localhost:8080")
    print("\nPresiona Ctrl+C para detener\n")
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n👋 Servidor detenido")
        sys.exit(0)
