FROM nginx:1.21
COPY nginx.conf /etc/nginx/nginx.conf
COPY index.html /usr/share/nginx/html/index.html
COPY selfsigned.crt /selfsigned.crt
COPY selfsigned.key /selfsigned.key

CMD ["nginx", "-g", "daemon off;"]
