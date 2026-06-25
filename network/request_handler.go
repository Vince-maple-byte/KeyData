package network

import (
	context "context"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	UnimplementedDataServer
}

func (s *Server) Search(ctx context.Context, search *SearchRequest) (*SearchResponse, error) {
	res, err := searchHelper(search.GetKey())

	if err != nil {
		return &SearchResponse{CreatedAt: nil, Key: "", Payload: ""}, status.Error(codes.NotFound, err.Error())
	}

	contents := record.GetContents(res)

	return &SearchResponse{
		CreatedAt: timestamppb.New(contents.Timestamp),
		Key:       contents.Key,
		Payload:   contents.Payload,
	}, nil
}

func (s *Server) Create(ctx context.Context, create *CreateRequest) (*CreateResponse, error) {
	mem := NetworkMemtable()

	ok, err := mem.Write(create.GetKey(), create.GetOperation(), create.Payload)

	if !ok {
		return &CreateResponse{Success: ok}, status.Error(codes.Unknown, err.Error())
	}

	return &CreateResponse{Success: ok}, nil
}

func searchHelper(key string) ([]byte, error) {
	mem := NetworkMemtable()

	res, err := mem.Get(key)

	if err != nil {
		return sstable.ReadFromAllFiles(key)
	}

	return res, err
}
