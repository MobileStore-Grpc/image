package service

import (
	mobilePb "github.com/MobileStore-Grpc/product/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Dial() (mobilePb.MobileServiceClient, error) {
	conn, err := grpc.Dial("172.17.0.2:8080", grpc.WithInsecure())
	if err != nil {
		return nil, logError(status.Errorf(codes.Internal, "cannot dial mobile search service", err))
	}
	mobileClient := mobilePb.NewMobileServiceClient(conn)
	return mobileClient, nil
}
