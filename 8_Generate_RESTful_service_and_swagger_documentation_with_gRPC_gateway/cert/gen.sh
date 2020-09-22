rm *.pem

# -------------------------- CA --------------------------
# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=CN/ST=GD/L=DZ/O=Ezzz/OU=Education/CN=*.ezzz.com/emailAddress=ezzz@gmail.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# -------------------------- 这里我使用相同的CA对server和client进行签名 --------------------------
# ------------------ 在真实环境中，我们可能有多个客户端，它们的证书由不同的CA签署 ------------------

# -------------------------- server --------------------------
# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=CN/ST=GD/L=DZ/O=Ezzz/OU=Education/CN=*.ezzz.com/emailAddress=ezzz@gmail.com"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

# -------------------------- client --------------------------
# 4. Generate web client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=CN/ST=GD/L=DZ/O=Pink/OU=Education/CN=*.pink.com/emailAddress=pink@gmail.com"

# 5. Use CA's private key to sign web client's CSR and get back the signed certificate
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf

echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text
