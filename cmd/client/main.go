package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/MobileStore-Grpc/image/pb"
	mobilePb "github.com/MobileStore-Grpc/product/pb"
	mobileSample "github.com/MobileStore-Grpc/product/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func uploadImage(imageClient pb.ImageServiceClient, mobileID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := imageClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{Info: &pb.ImageInfo{
			MobileId:  mobileID,
			ImageType: filepath.Ext(imagePath),
		}},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to the server: ", err, stream.RecvMsg(nil))
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err, stream.RecvMsg(nil))
	}

	log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}

func testUploadImage(mobileClient mobilePb.MobileServiceClient, imageClient pb.ImageServiceClient) {
	mobile := mobileSample.NewMobile()
	req := &mobilePb.CreateMobileRequest{
		Mobile: mobile,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := mobileClient.CreateMobile(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("mobile already exists")
		} else {
			log.Fatal("cannot create mobile: ", err)
		}
	}
	log.Printf("create mobile with id: %s", res.Id)
	uploadImage(imageClient, mobile.GetId(), "testdata/laptop.jpg")
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dail server %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial review service: ", err)
	}

	imageClient := pb.NewImageServiceClient(conn)
	mobileClient, err := Dial()
	if err != nil {
		log.Fatal("cannot dial mobile search service")
	}

	testUploadImage(mobileClient, imageClient)
}

func Dial() (mobilePb.MobileServiceClient, error) {
	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	mobileClient := mobilePb.NewMobileServiceClient(conn)
	return mobileClient, nil
}
