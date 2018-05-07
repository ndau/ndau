package ndau

// This file only exists to hold the generate command. We are using a special version of the
// protobuf compiler from gogo instead of google:
// https://github.com/gogo/protobuf
//
// We need to set the path at the project root so that we can import .proto files
// from our vendored dependencies. We also need to set it in the current directory;
// subdirectories are not searched.
//
//go:generate protoc --gogoslick_out=. transaction.proto --proto_path=. --proto_path=$GOPATH/src/gitlab.ndau.tech/experiments/ndau-chain/vendor
