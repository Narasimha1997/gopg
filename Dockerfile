FROM golang

COPY . /app
WORKDIR /app

RUN ./build.sh

ENTRYPOINT [ "./bin/gopg" ]
