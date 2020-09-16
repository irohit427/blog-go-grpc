package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/irohit427/go_grpc/blog/blog_pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct {
}

func dataToBlog(data *blogEntity) *blog_pb.Blog {
	return &blog_pb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (*server) CreateBlog(ctx context.Context, req *blog_pb.CreateBlogRequest) (*blog_pb.CreateBlogResponse, error) {
	blog := req.GetBlog()
	data := blogEntity{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}
	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Insert error: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert to OID"))
	}
	return &blog_pb.CreateBlogResponse{
		Blog: &blog_pb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, in *blog_pb.ReadBlogRequest) (*blog_pb.ReadBlogResponse, error) {
	fmt.Println("Read Blog Request")
	blogId := in.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}

	data := &blogEntity{}
	filter := bson.D{{"_id", oid}}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Blog not found"))
	}
	return &blog_pb.ReadBlogResponse{
		Blog: dataToBlog(data),
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, in *blog_pb.UpdateBlogRequest) (*blog_pb.UpdateBlogResponse, error) {
	fmt.Println("Update Blog Request")
	blog := in.GetBlog()
	oid, err := primitive.ObjectIDFromHex((blog.GetId()))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}

	data := &blogEntity{}
	filter := bson.D{{"_id", oid}}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Blog not found"))
	}

	data.AuthorID = blog.GetAuthorId()
	data.Title = blog.GetTitle()
	data.Content = blog.GetContent()

	_, err = collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot update object: %v", err))
	}

	return &blog_pb.UpdateBlogResponse{
		Blog: dataToBlog(data),
	}, nil
}

type blogEntity struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Connecting to MongoDB")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://rohit123:shlocked221b@cluster0.uqlap.mongodb.net/blogDB?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("blogDB").Collection("blog")
	fmt.Println("Connected to MongoDB")
	// Starting Server
	fmt.Println("Starting Server")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blog_pb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Ctrl+C to stop the server
	ch := make(chan os.Signal, 1)

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping server")
	s.Stop()
	fmt.Println("Closing Listener")
	lis.Close()
	fmt.Println("Closing MongoDB connection")
	client.Disconnect(context.Background())
	fmt.Println("Exit")
}
