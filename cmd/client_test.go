package cmd_test

import (
	"alice-tss/pb/tss"
	"context"
	"log"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCClient(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:2234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := tss.NewTssServiceClient(conn)

	bookList, err := client.SignMessage(context.Background(), &tss.SignRequest{
		Hash:    "0x5a73c8fb1b418fdd33985b0b3a8561243abbb5cf1af3f0a368502939e3a4d658",
		Pubkey:  "02d890e326fc2ea4f67d8eb6dc451779836fe7a15a2643b901d342f76ba06d7674",
		Message: "tss-service",
	})
	if err != nil {
		log.Fatalf("failed to get book list: %v", err)
	}
	log.Printf("book list: %v", bookList)
}
