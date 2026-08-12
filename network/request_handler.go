package network

import (
	context "context"
	"fmt"

	database "github.com/Vince-maple-byte/KeyData/internals/db"
	"github.com/Vince-maple-byte/KeyData/internals/record"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	UnimplementedDataServer

	Database *database.Database
}

func (s *Server) Search(ctx context.Context, search *SearchRequest) (*SearchResponse, error) {
	res, err := s.Database.Get(search.GetKey())

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
	//mem := NetworkMemtable()
	var ok bool
	var err error

	switch create.GetOperation() {
	case "PUT":
		ok, err = s.Database.Put(create.GetKey(), create.GetPayload())
	case "DELETE":
		ok, err = s.Database.Delete(create.GetKey())
	default:
		err = fmt.Errorf("invalid operation")
		ok = false
	}

	if !ok {
		return &CreateResponse{Success: ok}, status.Error(codes.Unknown, err.Error())
	}

	return &CreateResponse{Success: ok}, nil
}
