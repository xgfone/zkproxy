REVERSION=`git rev-parse HEAD`
VERSION=`git tag -l | tail -n 1`
go build -ldflags "-X main.reversion=${REVERSION} -X main.version=${VERSION}" github.com/xgfone/zkproxy
