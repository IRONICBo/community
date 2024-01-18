FROM golang:1.21.6

RUN apt-get update

WORKDIR /community

# install goplus
RUN cd .. && git clone https://github.com/goplus/gop.git && cd gop && ./all.bash

# run goplus-community
COPY . .
# RUN cd cmd/gopcomm && nohup gop run . >backend.log 2>&1 &
# RUN cd cmd/gopcomm/yap/community-frontend && npm install && nohup npm run dev >frontend.log 2>&1 &
CMD cd cmd/gopcomm && gop run .