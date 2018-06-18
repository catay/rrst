FROM golang:1.9

WORKDIR /go/src/app
ENV PATH="${PATH}:/go/src/app"

RUN git clone https://github.com/catay/rrst.git /go/src/app/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh 
RUN dep ensure -v
RUN go build -v -o rrst

CMD ["rrst"]


