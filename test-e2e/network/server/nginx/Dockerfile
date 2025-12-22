FROM nginx

COPY static /usr/share/nginx/html

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY configure.sh /configure.sh
RUN chmod +x /configure.sh

ENTRYPOINT ["bash"]
CMD ["/entrypoint.sh"]