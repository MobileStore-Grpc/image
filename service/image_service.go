package service

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/MobileStore-Grpc/image/pb"
	mobilePb "github.com/MobileStore-Grpc/product/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// maximum 1 megabyte
const maxImageSize = 1 << 20

//ImageService provides upload mobile images services
type ImageService struct {
	imageStore ImageStore
	pb.UnimplementedImageServiceServer
}

//NewReviewService returns a new image service
func NewImageService(imageStore ImageStore) *ImageService {
	return &ImageService{
		imageStore: imageStore,
	}
}

//UploadImage is a client-streaming RPC to upload a mobile Image
func (server *ImageService) UploadImage(stream pb.ImageService_UploadImageServer) error {
	//Creating connection to mobile service
	mobileClient, err := Dial()
	log.Print("dail connection to mobile service successfully")
	if err != nil {
		log.Println("cannot dial mobile search service")
		return err
	}

	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info: %v", err))
	}
	mobileID := req.GetInfo().GetMobileId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive an upload-image request for mobile %s with image type %s", mobileID, imageType)

	//check whether mobile exist or not before uploading image
	err = find(mobileID, mobileClient, stream.Context())
	if err != nil {
		return err
	}

	//upload image
	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		// check context error
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)
		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		//write slowly
		// time.Sleep(time.Second)

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chuk data: %v", err))
		}
	}
	imageID, err := server.imageStore.Save(mobileID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}
	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}
	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}
	log.Printf("save image with id: %s, size %d", imageID, imageSize)
	return nil
}

func find(mobileID string, mobileClient mobilePb.MobileServiceClient, ctx context.Context) error {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	req := mobilePb.SearchMobileRequest{
		MobileId: mobileID,
	}
	_, err := mobileClient.SearchMobile(ctx, &req)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	log.Printf("mobile with mobileID %s found", mobileID)
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
