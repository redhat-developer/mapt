package openshiftsnc

// This is the AWS policy required to use SSM service in order to set the values
// within userdata
var requiredPolicies = []string{"arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"}
