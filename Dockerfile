FROM circleci/golang:1.13

WORKDIR /go/src/github.com/catay/rrst
ENV PATH="${PATH}:/go/src/github.com/catay/rrst"
USER root

RUN git clone https://github.com/catay/rrst.git .
RUN make

CMD ["rrst"]
