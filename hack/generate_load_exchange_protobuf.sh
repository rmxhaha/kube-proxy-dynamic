cd pkg/load/exchange/
protoc -I loadexchange/ loadexchange.proto --go_out=plugins=grpc:loadexchange
