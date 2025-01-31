package constants

const (
	CONFIG_AWS_REGION        string = "aws:region"
	CONFIG_AWS_NATIVE_REGION string = "aws-native:region"
	CONFIG_AWS_ACCESS_KEY    string = "aws:accessKey"
	CONFIG_AWS_SECRET_KEY    string = "aws:secretKey"
)

const (
	MetadataBaseURL              = "http://169.254.170.2"
	ECSCredentialsRelativeURIENV = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
	DefaultAWSRegion             = "us-east-1"
)
