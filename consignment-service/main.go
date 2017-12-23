package main

import (
	"fmt"
	pb "github.com/austinsilver/ship/consignment-service/proto/consignment"
	vesselProto "github.com/austinsilver/ship/vesssel-service/proto/vessel"
	micro "github.com/micro/go-micro"
	"golang.org/x/net/context"
	"log"
)

type Respository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

// Repository - Dummy for future ues

type ConsignmentRepository struct {
	consignments []*pb.Consignment
}

func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

func (repo *ConsignmentRepository) GetAll() []*pb.Consignment {
	return repo.consignments
}

//Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition.  Check interface in generated code
type service struct {
	repo Respository
	vesselClient vesselProto.VesselServiceClient
}

//CreateConsignment one method on our service whic creates method which
// takes context and a requst as an argument
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

	//Call a client instance of our vessel service with our consignment weight, and amount of the containsers as capacity
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specificataion{
		MaxWeight: req.Weight,
		Capacity: int32(len(req.Containers)),
	})

	log.Panicf("Found Vessel: %s \n", vesselResponse.Vessel.Name)
	if err != nil {
		return err
	}

	req.VesselId = vesselResponse.Vessel.Id

	//Save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}

	// return maatching 'Response'
	res.Created = true
	res.Consignment = consignment
	return nil

}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments := s.repo.GetAll()
	res.Consignments = consignments
	return nil
}

func main() {
	repo := &ConsignmentRepository{}

	//Setup a new service.

	srv := micro.NewService(
		// Name must match the package name givin
		micro.Name("go.micro.srv.consignment"),
		micro.Version("latest"),
	)

	vesselClient := vesselProto.NewVesselServiceClient("go.micro.srv.vessel", srv.Client())
	//Init will parse the command line flags
	srv.Init()

	//Register the Service with gRPC server and tie our autogen code
	// interface code for our protobuf
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})

	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
