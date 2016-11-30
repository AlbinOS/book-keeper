FROM centurylink/ca-certs
MAINTAINER Albin Gilles <gilles.albin@gmail.com>
ENV REFRESHED_AT 2016-11-30

EXPOSE 8080

COPY ./book-keeper /.

ENTRYPOINT ["/book-keeper"]
CMD ["--help"]
