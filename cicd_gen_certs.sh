#This is for CICD workflow that has the key pulled from the secrets repo and store as an env var
#Generates server keys and signs them with the CA
base64 -d myca.c.b64 >myCA.pem
echo $CA_KEY | base64 -d > myCA.key
openssl genrsa -out out/k8srelay.key 2048
openssl req -new -key out/k8srelay.key -out out/k8srelay.csr  -subj /C=US/O=theOTW/OU=Engineering/CN=k8srelay
openssl x509 -req -in out/k8srelay.csr -CA myCA.pem -CAkey myCA.key -CAcreateserial -out out/k8srelay.crt -days 825 -sha256 -extfile samplecerts/x509.config

rm  myCA.pem
