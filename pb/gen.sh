#cd .. && protoc -I ./protobuf/ ./protobuf/builder.proto --go_out=plugins=grpc:builder
go get github.com/golang/protobuf/{proto,protoc-gen-go}
protoc --go_out=plugins=grpc:../protobuf *.proto
