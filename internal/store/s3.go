package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/opts"
)

// BucketPolicyVersion is the version of the S3 bucket policy.
// This is a constant that should not change.
const BucketPolicyVersion = "2012-10-17"

type S3BucketPolicy struct {
	Version   string                    `json:"Version"`
	Statement []S3BucketPolicyStatement `json:"Statement"`
}

type S3BucketPolicyStatement struct {
	Sid       string        `json:"Sid"`
	Effect    string        `json:"Effect"`
	Principal Principal     `json:"Principal"`
	Action    StringOrSlice `json:"Action"`
	Resource  StringOrSlice `json:"Resource"`
}

func (s *S3BucketPolicyStatement) IsEqual(other S3BucketPolicyStatement) bool {
	return s.Sid == other.Sid &&
		s.Effect == other.Effect &&
		s.Principal.IsEqual(other.Principal) &&
		slices.Equal(s.Action, other.Action) && // Compare Action as slices
		slices.Equal(s.Resource, other.Resource) // Compare Resource as slices
}

func DesiredBucketPolicyStatement(bucketName string) S3BucketPolicyStatement {
	var resource StringOrSlice

	switch config.C.Stores.S3.Type {
	// case "seaweedfs":
	// 	resource = StringOrSlice{fmt.Sprintf("%s/public/*", bucketName)}
	default:
		resource = StringOrSlice{fmt.Sprintf("arn:aws:s3:::%s/public/*", bucketName)}
	}
	return S3BucketPolicyStatement{
		Sid:       "PublicReadGetObject",
		Effect:    "Allow",
		Principal: Principal{Str: "*", IsStr: true},
		Action:    StringOrSlice{"s3:GetObject"},
		Resource:  resource,
	}
}

// Principal type to handle string or map[string]string
type Principal struct {
	Map   map[string][]string
	Str   string
	IsStr bool
}

func (p *Principal) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		p.Str = single
		p.IsStr = true
		return nil
	}
	var multi map[string][]string
	if err := json.Unmarshal(data, &multi); err == nil {
		p.Map = multi
		return nil
	}
	return fmt.Errorf("invalid Principal format %v", string(data))
}

func (p *Principal) MarshalJSON() ([]byte, error) {
	if p.IsStr {
		return json.Marshal(p.Str)
	}
	return json.Marshal(p.Map)
}

func (p Principal) IsEqual(other Principal) bool {
	if p.IsStr != other.IsStr {
		return false
	}
	if p.IsStr && other.IsStr {
		return p.Str == other.Str
	}
	if len(p.Map) != len(other.Map) {
		return false
	}
	for k, v := range p.Map {
		if otherV, ok := other.Map[k]; !ok || !slices.Equal(v, otherV) {
			return false
		}
	}
	return true
}

// StringOrSlice handles "foo" or ["foo", "bar"]
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}
	var multi []string
	if err := json.Unmarshal(data, &multi); err == nil {
		*s = multi
		return nil
	}
	return fmt.Errorf("invalid Action/Resource format")
}

func (s *StringOrSlice) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(*s))
}

func NewS3Store() *S3Store {
	cfg := aws.Config{
		Region: config.C.Stores.S3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			config.C.Stores.S3.AccessKeyID,
			config.C.Stores.S3.SecretAccessKey,
			"",
		),
	}
	// If the endpoint is set, use it; otherwise, use the default AWS endpoint.
	scheme := "http"
	if config.C.Stores.S3.UseSSL {
		scheme = "https"
	}
	return &S3Store{
		Client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			if strings.TrimSpace(config.C.Stores.S3.Endpoint) != "" {
				endpoint := scheme + "://" + strings.TrimSuffix(config.C.Stores.S3.Endpoint, "/")
				o.BaseEndpoint = &endpoint
				o.UsePathStyle = true // Enable path-style URLs for non S3 endpoints
			}
		}),
		Bucket:         aws.String(opts.S3StoreBucketName),
		EndpointScheme: scheme,
	}
}

// IsErrorCode checks whether err is a smithy.APIError with one of the given codes.
// Falls back to substring match if exact match doesn't work (for MinIO/alt-S3).
func (s *S3Store) isErrorCode(err error, codes ...string) bool {
	if err == nil {
		return false
	}

	var apiErr smithy.APIError
	if ok := errors.As(err, &apiErr); ok {
		for _, code := range codes {
			if apiErr.ErrorCode() == code || strings.Contains(apiErr.ErrorCode(), code) {
				return true
			}
		}
	}
	return false
}

type S3Store struct {
	Client *s3.Client
	Bucket *string
	// Used to construct the endpoint URL for objects.
	// "http" or "https"
	EndpointScheme string
}

// Init initializes the S3Store.
//
// We do things like setting bucket policies, creating buckets, etc.
// This method can be extended to perform any necessary setup for the S3 store.
// This is stateless, and only needs to be called once during the application startup.
func (s *S3Store) Init(ctx context.Context) error {
	// Desired bucket policy statement
	desiredPolicyStatement := DesiredBucketPolicyStatement(*s.Bucket)

	// Initialization logic if needed
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: s.Bucket,
	})

	if err != nil {
		if s.isErrorCode(err, "NotFound", "NoSuchBucket") {
			// If the bucket does not exist, we can create it.
			_, err = s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: s.Bucket,
			})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", *s.Bucket, err)
			}
		} else {
			return fmt.Errorf("failed to check bucket existence: %w", err)
		}
	}

	// if no error, the bucket exists

	// Check for bucket policy to ensure public/* access is allowed publicly.
	bucketPolicyResult, err := s.Client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: s.Bucket,
	})

	if err != nil {
		if s.isErrorCode(err, "NoSuchBucketPolicy") {
			// If the bucket policy does not exist, we can create it to allow public access to the bucket.
			desiredBucketPolicy := S3BucketPolicy{
				Version: BucketPolicyVersion,
				Statement: []S3BucketPolicyStatement{
					desiredPolicyStatement,
				},
			}

			policyBytes, err := json.Marshal(&desiredBucketPolicy)
			if err != nil {
				return fmt.Errorf("failed to marshal desired bucket policy: %w", err)
			}

			_, err = s.Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
				Bucket: s.Bucket,
				Policy: aws.String(string(policyBytes)),
			})

			if err != nil {
				return fmt.Errorf("failed to put bucket policy: %w", err)
			}
			// Bucket policy created successfully and bucket exists, return nil.
			return nil
		} else {
			return fmt.Errorf("failed to get bucket policy: %w", err)
		}
	}

	bucketPolicy := S3BucketPolicy{}

	err = json.Unmarshal([]byte(*bucketPolicyResult.Policy), &bucketPolicy)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bucket policy: %w", err)
	}

	for _, statement := range bucketPolicy.Statement {
		if statement.IsEqual(desiredPolicyStatement) {
			// The desired bucket policy already exists, no need to create it again.
			return nil
		}
	}

	// If we reach here, it means the desired bucket policy does not exist.
	// We can put the desired policy.

	bucketPolicy.Statement = append(bucketPolicy.Statement, desiredPolicyStatement)
	policyBytes, err := json.Marshal(&bucketPolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal updated bucket policy: %w", err)
	}
	_, err = s.Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: s.Bucket,
		Policy: aws.String(string(policyBytes)),
	})
	if err != nil {
		if s.isErrorCode(err, "NoSuchBucketPolicy") {

			// If the error is "NoSuchBucketPolicy", it means the policy does not exist.
			// We can return an error indicating that the desired policy does not exist.
			// This is a fatal error, as the bucket policy needs to be created manually or by calling this method again (on restart).
			return fmt.Errorf("desired bucket policy does not exist, auto creation failed, please create it manually: %s, or restart this instance", desiredPolicyStatement.Sid)
		} else {
			return fmt.Errorf("failed to put bucket policy: %w", err)
		}
	}

	// Bucket policy updated successfully, return nil.
	return nil
}

func (s *S3Store) GetBucketName() string {
	if s.Bucket == nil {
		return ""
	}
	return *s.Bucket
}

func (s *S3Store) uploadObjectInternal(ctx context.Context, key string, file io.Reader, contentType, cacheControl string) (string, error) {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       s.Bucket,
		Key:          aws.String(key),
		Body:         file,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(cacheControl),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload object to bucket %s with key %s: %w", *s.Bucket, key, err)
	}

	return key, nil
}

func (s *S3Store) UploadObject(ctx context.Context, key string, file io.Reader, contentType, cacheControl string) (string, error) {
	return s.uploadObjectInternal(ctx, "private/"+key, file, contentType, cacheControl)
}

func (s *S3Store) UploadPublicObject(ctx context.Context, key string, file io.Reader, contentType, cacheControl string) (string, error) {
	return s.uploadObjectInternal(ctx, "public/"+key, file, contentType, cacheControl)
}

// DeleteObject deletes an object with the given key from the S3 bucket.
// The key should be prefixed with 'private/' or 'public/' as per the upload methods.
func (s *S3Store) DeleteObject(ctx context.Context, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: s.Bucket,
		Key:    aws.String(key),
	})
	return err
}

// GetObjectURL returns the URL of an object with the given key.
// The key should be prefixed with 'private/' or 'public/' as per the upload methods.
// If the bucket is public, it returns a URL that can be accessed directly.
// If the bucket is private, it returns a URL that can be used to access the object
// using a pre-signed URL (not implemented here).
// Note: This method does not return a pre-signed URL, so it may not be accessible if the object is private.
func (s *S3Store) GetObjectURL(ctx context.Context, key string) (string, error) {
	// For stores other than AWS S3, we can return a path style URL.
	if config.C.Stores.S3.Endpoint != "" {
		return fmt.Sprintf("%s://%s/%s/%s", s.EndpointScheme, strings.TrimSuffix(config.C.Stores.S3.Endpoint, "/"), *s.Bucket, key), nil
	}
	// For the public file access, we need to prepend the bucket name to the endpoint.
	// like "https://<bucket-name>.s3.<region>.amazonaws.com"
	return fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s", s.EndpointScheme, *s.Bucket, config.C.Stores.S3.Region, key), nil
}

// STUB, we just return a value using GetObjectURL
// TODO: Implement this method to return a pre-signed URL for the object with expiry.
func (s *S3Store) GetObjectURLWithExpiry(ctx context.Context, key string, expiry int64) (string, error) {
	return s.GetObjectURL(ctx, key)
}
