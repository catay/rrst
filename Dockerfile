FROM golang:1.9

WORKDIR /go/src/rrst
ENV PATH="${PATH}:/go/src/rrst"

RUN git clone https://github.com/catay/rrst.git /go/src/rrst/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh 
RUN make

CMD ["rrst"]


