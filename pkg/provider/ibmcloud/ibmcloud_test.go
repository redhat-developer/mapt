package ibmcloud

import "testing"

func TestIsCOSBackend(t *testing.T) {
	for _, u := range []string{"cos://bucket/path", "s3://bucket/path", "https://s3.us-south.cloud-object-storage.appdomain.cloud/bucket"} {
		if !isCOSBackend(u) {
			t.Errorf("%q should be a COS backend", u)
		}
	}
	if isCOSBackend("azblob://container") {
		t.Error("azblob should not be a COS backend")
	}
}

func TestExtractBucketAndPathCOSScheme(t *testing.T) {
	bucket, path, err := extractBucketAndPath("cos://my-bucket/some/path")
	if err != nil {
		t.Fatal(err)
	}
	if bucket != "my-bucket" || path != "some/path" {
		t.Errorf("got %q %q", bucket, path)
	}
	if _, _, err := extractBucketAndPath("cos://"); err == nil {
		t.Error("expected an error for a missing bucket")
	}
}

func TestInitCOSBackendFromCOSScheme(t *testing.T) {
	t.Setenv("IBMCLOUD_COS_ACCESS_KEY_ID", "k")
	t.Setenv("IBMCLOUD_COS_SECRET_ACCESS_KEY", "s")
	t.Setenv("IC_REGION", "us-south")
	t.Setenv("IBMCLOUD_COS_ENDPOINT", "")
	got, err := initCOSBackend("cos://bucket/some/path")
	if err != nil {
		t.Fatal(err)
	}
	want := "s3://bucket/some/path?endpoint=https://s3.us-south.cloud-object-storage.appdomain.cloud&s3ForcePathStyle=true"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
