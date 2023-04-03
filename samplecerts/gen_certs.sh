#Caution, making a new CA will invalidate client configs, that is what the next 4 lines do
openssl genrsa -out myCA.key 2048
openssl req -x509 -new -nodes -key myCA.key -sha256 -days 1825 -out myCA.pem -subj /C=US/O=theOTW/OU=Engineering/CN=k8sca
base64 -i myCA.key -o myca.k.b64
base64 -i myCA.pem -o myca.c.b64

#Generates server keys and signs them with the CA
base64 -d -o myCA.pem -i myca.c.b64
base64 -d -o myCA.key -i myca.k.b64
openssl genrsa -out k8srelay.key 2048
openssl req -new -key k8srelay.key -out k8srelay.csr  -subj /C=US/O=theOTW/OU=Engineering/CN=k8srelay
openssl x509 -req -in k8srelay.csr -CA myCA.pem -CAkey myCA.key -CAcreateserial -out k8srelay.crt -days 825 -sha256 -extfile x509.config
rm myCA.pem
rm myCA.key

