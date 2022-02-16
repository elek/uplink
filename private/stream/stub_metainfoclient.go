package stream

import (
	"context"
	"fmt"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/drpc"
)

type metaInfoClientStub struct {
	nodes stubNodes
}

func (m metaInfoClientStub) DRPCConn() drpc.Conn {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) CreateBucket(ctx context.Context, in *pb.BucketCreateRequest) (*pb.BucketCreateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) GetBucket(ctx context.Context, in *pb.BucketGetRequest) (*pb.BucketGetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) DeleteBucket(ctx context.Context, in *pb.BucketDeleteRequest) (*pb.BucketDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) ListBuckets(ctx context.Context, in *pb.BucketListRequest) (*pb.BucketListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) BeginObject(ctx context.Context, in *pb.ObjectBeginRequest) (*pb.ObjectBeginResponse, error) {
	fmt.Println("begin object")
	return &pb.ObjectBeginResponse{
		Bucket:        in.Bucket,
		EncryptedPath: in.EncryptedPath,
	}, nil
}

func (m metaInfoClientStub) CommitObject(ctx context.Context, in *pb.ObjectCommitRequest) (*pb.ObjectCommitResponse, error) {
	fmt.Println("commit object")
	return &pb.ObjectCommitResponse{}, nil
}

func (m metaInfoClientStub) GetObject(ctx context.Context, in *pb.ObjectGetRequest) (*pb.ObjectGetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) GetPendingObjects(ctx context.Context, in *pb.GetPendingObjectsRequest) (*pb.GetPendingObjectsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) ListObjects(ctx context.Context, in *pb.ObjectListRequest) (*pb.ObjectListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) BeginDeleteObject(ctx context.Context, in *pb.ObjectBeginDeleteRequest) (*pb.ObjectBeginDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) FinishDeleteObject(ctx context.Context, in *pb.ObjectFinishDeleteRequest) (*pb.ObjectFinishDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) GetObjectIPs(ctx context.Context, in *pb.ObjectGetIPsRequest) (*pb.ObjectGetIPsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) ListPendingObjectStreams(ctx context.Context, in *pb.ObjectListPendingStreamsRequest) (*pb.ObjectListPendingStreamsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) DownloadObject(ctx context.Context, in *pb.ObjectDownloadRequest) (*pb.ObjectDownloadResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) UpdateObjectMetadata(ctx context.Context, in *pb.ObjectUpdateMetadataRequest) (*pb.ObjectUpdateMetadataResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) BeginSegment(ctx context.Context, in *pb.SegmentBeginRequest) (*pb.SegmentBeginResponse, error) {
	fmt.Println("begin segment")
	limits := []*pb.AddressedOrderLimit{}
	for i := 0; i < 20; i++ {
		limits = append(limits, &pb.AddressedOrderLimit{
			Limit: &pb.OrderLimit{
				Limit: 1024,
			},
			StorageNodeAddress: &pb.NodeAddress{
				Transport: pb.NodeTransport_TCP_TLS_GRPC,
				Address:   m.nodes[i].Address,
			},
		})
	}

	_, key, err := storj.NewPieceKey()
	if err != nil {
		return nil, err
	}
	return &pb.SegmentBeginResponse{
		SegmentId:       pb.SegmentID{},
		AddressedLimits: limits,
		PrivateKey:      key,
		RedundancyScheme: &pb.RedundancyScheme{
			Type:             pb.RedundancyScheme_RS,
			MinReq:           10,
			Total:            20,
			RepairThreshold:  11,
			SuccessThreshold: 20,
			ErasureShareSize: 1024,
		},
	}, nil
}

func (m metaInfoClientStub) CommitSegment(ctx context.Context, in *pb.SegmentCommitRequest) (*pb.SegmentCommitResponse, error) {
	return &pb.SegmentCommitResponse{
		SuccessfulPieces: int32(len(in.UploadResult)),
	}, nil
}

func (m metaInfoClientStub) MakeInlineSegment(ctx context.Context, in *pb.SegmentMakeInlineRequest) (*pb.SegmentMakeInlineResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) BeginDeleteSegment(ctx context.Context, in *pb.SegmentBeginDeleteRequest) (*pb.SegmentBeginDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) FinishDeleteSegment(ctx context.Context, in *pb.SegmentFinishDeleteRequest) (*pb.SegmentFinishDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) ListSegments(ctx context.Context, in *pb.SegmentListRequest) (*pb.SegmentListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) DownloadSegment(ctx context.Context, in *pb.SegmentDownloadRequest) (*pb.SegmentDownloadResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) DeletePart(ctx context.Context, in *pb.PartDeleteRequest) (*pb.PartDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) Batch(ctx context.Context, in *pb.BatchRequest) (*pb.BatchResponse, error) {
	response := &pb.BatchResponse{
		Responses: []*pb.BatchResponseItem{},
	}
	for _, r := range in.GetRequests() {
		switch e := r.Request.(type) {
		case *pb.BatchRequestItem_ObjectBegin:
			resp, err := m.BeginObject(ctx, e.ObjectBegin)
			if err != nil {
				return nil, err
			}
			response.Responses = append(response.Responses, &pb.BatchResponseItem{
				Response: &pb.BatchResponseItem_ObjectBegin{
					ObjectBegin: resp,
				},
			})
		case *pb.BatchRequestItem_SegmentBegin:
			resp, err := m.BeginSegment(ctx, e.SegmentBegin)
			if err != nil {
				return nil, err
			}
			response.Responses = append(response.Responses, &pb.BatchResponseItem{
				Response: &pb.BatchResponseItem_SegmentBegin{
					SegmentBegin: resp,
				},
			})
		case *pb.BatchRequestItem_SegmentCommit:
			resp, err := m.CommitSegment(ctx, e.SegmentCommit)
			if err != nil {
				return nil, err
			}
			response.Responses = append(response.Responses, &pb.BatchResponseItem{
				Response: &pb.BatchResponseItem_SegmentCommit{
					SegmentCommit: resp,
				},
			})
		case *pb.BatchRequestItem_ObjectCommit:
			resp, err := m.CommitObject(ctx, e.ObjectCommit)
			if err != nil {
				return nil, err
			}
			response.Responses = append(response.Responses, &pb.BatchResponseItem{
				Response: &pb.BatchResponseItem_ObjectCommit{
					ObjectCommit: resp,
				},
			})
		default:
			panic(fmt.Sprintf("%T is not supported", e))
		}
	}
	return response, nil
}

func (m metaInfoClientStub) ProjectInfo(ctx context.Context, in *pb.ProjectInfoRequest) (*pb.ProjectInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) RevokeAPIKey(ctx context.Context, in *pb.RevokeAPIKeyRequest) (*pb.RevokeAPIKeyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) BeginMoveObject(ctx context.Context, in *pb.ObjectBeginMoveRequest) (*pb.ObjectBeginMoveResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m metaInfoClientStub) FinishMoveObject(ctx context.Context, in *pb.ObjectFinishMoveRequest) (*pb.ObjectFinishMoveResponse, error) {
	//TODO implement me
	panic("implement me")
}

var _ pb.DRPCMetainfoClient = &metaInfoClientStub{}
