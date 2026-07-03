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

	Memtable InternalMemtable
	DataDir  string
}

func (s *Server) Search(ctx context.Context, search *SearchRequest) (*SearchResponse, error) {
	res, err := s.searchHelper(search.GetKey())

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

	ok, err := s.Memtable.Write(create.GetKey(), create.GetPayload(), create.GetOperation())

	if !ok {
		return &CreateResponse{Success: ok}, status.Error(codes.Unknown, err.Error())
	}

	return &CreateResponse{Success: ok}, nil
}

func (s Server) searchHelper(key string) ([]byte, error) {
	//mem := NetworkMemtable()

	res, err := s.Memtable.Get(key)

	//Got to figure out how to pass in the filePath for the directory while it still being reusable
	if err != nil {
		return sstable.ReadFromAllFiles(key, s.DataDir)
	}

	return res, err
}
