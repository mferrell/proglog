package http

import (
	"github.com/gogo/gateway"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/gogo/grpc-example/insecure"
	pbExample "github.com/gogo/grpc-example/proto"
	"github.com/gogo/grpc-example/server"
)
