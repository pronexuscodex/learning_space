# ECHO -- a server and a client talking over a real TCP socket.
import socket
import threading

server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.bind(("127.0.0.1", 0))            # port 0: let the OS choose a free port
server.listen(1)
port = server.getsockname()[1]

def serve():
    conn, _ = server.accept()
    data = conn.recv(1024)
    conn.sendall(b"echo: " + data.upper())
    conn.close()

threading.Thread(target=serve).start()

client = socket.create_connection(("127.0.0.1", port))
client.sendall(b"hello, 1985")
print(client.recv(1024).decode())
client.close()
server.close()
