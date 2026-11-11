package order

import "learn-grpc-gql-micro/account"

type grpcServer struct {
	service       Service
	accountClient *account.Client
}

