#!/bin/sh

rm ./*.pb.go >> /dev/null
rm ./api/*.pb.go >> /dev/null
protoc -I=./api -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
--gogofaster_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types:. \
 api/avalonGame.proto
sed -i -E 's/json:("[^"]+,omitempty")/json:\1 bson:\1/' ./*.pb.go
mv ./*.pb.go ./api
