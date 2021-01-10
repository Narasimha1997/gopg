FROM golang

RUN apt-get update -y && \
    apt-get install -qy curl && \
    apt-get install -qy curl && \
    curl -sSL https://get.docker.com/ | sh

#install gVisor + runsc
RUN apt-get update -y && apt install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

#add ppa
RUN curl -fsSL https://gvisor.dev/archive.key | apt-key add - && \
    add-apt-repository "deb https://storage.googleapis.com/gvisor/releases release main"

RUN apt-get update && apt-get install -y runsc

#set Runsc as docker runtime
RUN /usr/bin/runsc install

WORKDIR /

COPY src /app

#build the go app
RUN cd /app && go build -o /bin/gopg && cd ..
RUN rm -r /app

ENTRYPOINT ["/bin/gopg"]


