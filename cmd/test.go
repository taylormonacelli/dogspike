/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test called")
		test()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Define a struct to hold bucket information
type BucketInfo struct {
	Name      string
	ItemCount int64
	TotalSize int64
}

func test() {
	// Create a new AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		fmt.Println("Failed to load AWS configuration:", err)
		return
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Retrieve the list of S3 buckets in the specified region
	resp, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		fmt.Println("Failed to retrieve S3 buckets:", err)
		return
	}

	// Create a slice to hold the bucket information
	bucketList := make([]BucketInfo, 0)

	// Iterate over each bucket
	for _, bucket := range resp.Buckets {
		// Create the input for the GetBucketLocation API
		input := &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		}

		// Call the GetBucketLocation API to retrieve the bucket's region
		output, err := s3Client.GetBucketLocation(context.TODO(), input)
		if err != nil {
			panic("failed to get bucket location: " + err.Error())
		}

		// Retrieve the region from the output
		region := output.LocationConstraint

		// Generate the endpoint URL using the region and bucket name
		endpoint := fmt.Sprintf("https://s3.%s.amazonaws.com/%s", region, *bucket.Name)

		fmt.Println("Bucket Endpoint:", endpoint)

		// Create an S3 client with the custom endpoint
		client := s3.NewFromConfig(cfg, func(options *s3.Options) {
			options.EndpointResolver = s3.EndpointResolverFromURL(endpoint)
		})

		// Create the input for ListObjectsV2 operation
		input2 := &s3.ListObjectsV2Input{
			Bucket: bucket.Name,
		}

		// Retrieve the objects in the bucket
		objectResp, err := client.ListObjectsV2(context.TODO(), input2)
		if err != nil {
			fmt.Printf("Failed to retrieve objects for bucket '%s': %v\n", *bucket.Name, err)
			continue
		}

		// Calculate the summary information
		totalSize := int64(0)
		totalObjects := int64(0)

		for _, obj := range objectResp.Contents {
			totalSize += obj.Size
			totalObjects++
		}

		// Create a BucketInfo struct and add it to the bucketList
		bucketInfo := BucketInfo{
			Name:      *bucket.Name,
			ItemCount: totalObjects,
			TotalSize: totalSize,
		}

		bucketList = append(bucketList, bucketInfo)

		// Print the bucket information
		for _, bucketInfo := range bucketList {
			fmt.Printf("Bucket Name: %s\n", bucketInfo.Name)
			fmt.Printf("Item Count: %d\n", bucketInfo.ItemCount)
			fmt.Printf("Total Size: %s\n", formatSize(bucketInfo.TotalSize))
			fmt.Println()
		}
	}
}

// Helper function to format the size in a human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
