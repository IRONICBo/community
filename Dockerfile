FROM golang:1.21.6

RUN apt-get update

WORKDIR /community

# install goplus
RUN cd .. && git clone https://github.com/goplus/gop.git && cd gop && ./all.bash

# run goplus-community
COPY . .
CMD bash scripts/start.sh
