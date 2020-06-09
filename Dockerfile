FROM golang
ADD . /go/src/github.com/branthz/etcd
RUN go install github.com/branthz/etcd
EXPOSE 2379 2380
ENTRYPOINT ["etcd"]
