FROM nginx:1.27-alpine

COPY services/gateway/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY services/gateway/docs /usr/share/nginx/html/docs

EXPOSE 8084
