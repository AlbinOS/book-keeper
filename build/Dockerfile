FROM centurylink/ca-certs
MAINTAINER Albin Gilles <gilles.albin@gmail.com>
ENV REFRESHED_AT 2016-11-14

EXPOSE 8080

COPY ./book-keeper /.
COPY ./static /static
COPY ./views /views

ENTRYPOINT ["/book-keeper"]
CMD ["--help"]
