package main

import (
	"context"
	"fmt"
	"log"

	"github.com/irohit427/go_grpc/blog/blog_pb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Blog Client")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Couldn't connect %v", err)
	}
	defer cc.Close()

	c := blog_pb.NewBlogServiceClient(cc)

	blog := &blog_pb.Blog{
		AuthorId: "irohit427",
		Title:    "First Blog",
		Content:  "This is the first blog",
	}

	res, err := c.CreateBlog(context.Background(), &blog_pb.CreateBlogRequest{Blog: blog})

	if err != nil {
		log.Fatalf("Failed to create blog %v", err)
	}

	fmt.Println("Blog : %v", res)

	// read Blog
	res2, err := c.ReadBlog(context.Background(), &blog_pb.ReadBlogRequest{BlogId: "5f611acf6048a75681eb10df"})
	if err != nil {
		fmt.Printf("Err happened while reading: %v", err)
	}

	fmt.Println("Blog: %v", res2)

	// update blog
	blog = &blog_pb.Blog{
		Id:       "5f611acf6048a75681eb10df",
		AuthorId: "irohit427",
		Title:    "First Blog",
		Content:  "I'm modifying the content of this blog",
	}
	updateRes, err := c.UpdateBlog(context.Background(), &blog_pb.UpdateBlogRequest{Blog: blog})
	if err != nil {
		fmt.Printf("Unable to update blog, Something went wrong: %v", err)
	}
	fmt.Println("Updated Blog:", updateRes)

	// Delete Blog
	deleteRes, err := c.DeleteBlog(context.Background(), &blog_pb.DeleteBlogRequest{BlogId: "5f6264f5bea6d910a9eccf68"})
	if err != nil {
		fmt.Printf("Unable to delete blog, Something went wrong: %v", err)
	}
	fmt.Println("Deleted Blog:", deleteRes)
}
