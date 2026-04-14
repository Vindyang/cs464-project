package s3handler

import (
	"testing"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestValidateRegion(t *testing.T) {
	t.Run("accepts standard AWS region", func(t *testing.T) {
		if err := ValidateRegion("ap-southeast-1"); err != nil {
			t.Fatalf("ValidateRegion returned error: %v", err)
		}
	})

	t.Run("rejects malformed region", func(t *testing.T) {
		if err := ValidateRegion("Sydney"); err == nil {
			t.Fatal("ValidateRegion returned nil, want error")
		}
	})
}

func TestBucketNameForAccount(t *testing.T) {
	got, err := bucketNameForAccount("027742773650")
	if err != nil {
		t.Fatalf("bucketNameForAccount returned error: %v", err)
	}

	want := "omnishard-027742773650"
	if got != want {
		t.Fatalf("bucket name: got %q want %q", got, want)
	}
}

func TestIsValidBucketName(t *testing.T) {
	tests := []struct {
		name   string
		bucket string
		want   bool
	}{
		{name: "valid lowercase bucket", bucket: "omnishard-027742773650", want: true},
		{name: "rejects uppercase", bucket: "Omnishard-027742773650", want: false},
		{name: "rejects underscores", bucket: "omnishard_027742773650", want: false},
		{name: "rejects leading hyphen", bucket: "-omnishard-027742773650", want: false},
		{name: "rejects trailing hyphen", bucket: "omnishard-027742773650-", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidBucketName(test.bucket); got != test.want {
				t.Fatalf("isValidBucketName(%q): got %t want %t", test.bucket, got, test.want)
			}
		})
	}
}

func TestCreateBucketInput(t *testing.T) {
	t.Run("omits location constraint for us-east-1", func(t *testing.T) {
		input := createBucketInput("omnishard-027742773650", "us-east-1")
		if input.CreateBucketConfiguration != nil {
			t.Fatal("CreateBucketConfiguration should be nil for us-east-1")
		}
	})

	t.Run("sets location constraint for non-default region", func(t *testing.T) {
		input := createBucketInput("omnishard-027742773650", "ap-southeast-1")
		if input.CreateBucketConfiguration == nil {
			t.Fatal("CreateBucketConfiguration should not be nil")
		}
		if got := input.CreateBucketConfiguration.LocationConstraint; got != s3types.BucketLocationConstraint("ap-southeast-1") {
			t.Fatalf("location constraint: got %q want %q", got, s3types.BucketLocationConstraint("ap-southeast-1"))
		}
	})
}
