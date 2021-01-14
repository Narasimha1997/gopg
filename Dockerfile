FROM golang

RUN apt-get update -y && \
    apt-get install -qy curl && \
    apt-get install -qy curl && \
    curl -sSL https://get.docker.com/ | sh

WORKDIR /

COPY src /app

#build the go app
RUN cd /app && go build -o /bin/gopg && cd ..
RUN rm -r /app

ENTRYPOINT ["/bin/gopg"]


