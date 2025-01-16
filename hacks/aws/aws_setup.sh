#!/bin/bash

#Usage ./../aws_setup.sh ACCESS_KEY SECRET_KEY REGION TEAM_ID PROJECT_ID

CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-"podman"}"
AWS_CLI="${CONTAINER_RUNTIME} run --rm -it -e AWS_ACCESS_KEY_ID=${1} -e AWS_SECRET_ACCESS_KEY=${2} -e AWS_DEFAULT_REGION=${3} docker.io/amazon/aws-cli:latest"

aws_cmd () {
    ${AWS_CLI} ${1}
}

jq_cmd () {
    echo '#!/bin/bash' | tee tmp-jq.sh>/dev/null
    echo "jq ${2} /data" | tee -a tmp-jq.sh>/dev/null
    chmod +x tmp-jq.sh
    result=$(${CONTAINER_RUNTIME} run -v "$PWD/${1}":/data:Z -v "$PWD/tmp-jq.sh":/usr/local/bin/tmp-jq.sh:Z -ti quay.io/biocontainers/jq:1.6 tmp-jq.sh)>/dev/null
    rm tmp-jq.sh
    echo $result
}

# Create a group 
group_name="${4}-infra-management-${5}"
aws_cmd "iam get-group --group-name ${group_name}"
if [[ $? -ne 0 ]]; then 
    aws_cmd "iam create-group --group-name ${group_name}"
fi

# Create user
user_name=${4}-${5}
aws_cmd "iam get-user --user-name ${user_name}"
if [[ $? -ne 0 ]]; then 
    aws_cmd "iam create-user --user-name ${user_name}"
    aws_cmd "iam create-access-key --user-name ${user_name}" > access_key_info
fi

# Add user to group
aws_cmd "iam add-user-to-group --user-name ${user_name} --group-name ${group_name}"

# Got creds for user / sa
access_key=$(jq_cmd access_key_info "'.AccessKey.AccessKeyId'")>/dev/null
secret_key=$(jq_cmd access_key_info "'.AccessKey.SecretAccessKey'")>/dev/null
echo "ACCESS KEY: ${access_key}"
echo "SECRET KEY: ${secret_key}"

# Add policies to group
# Policies
AmazonVPCFullAccess_arn="arn:aws:iam::aws:policy/AmazonVPCFullAccess"
aws_cmd "iam attach-group-policy --group-name ${group_name} --policy-arn ${AmazonVPCFullAccess_arn}"
AmazonEC2FullAccess_arn="arn:aws:iam::aws:policy/AmazonEC2FullAccess"
aws_cmd "iam attach-group-policy --group-name ${group_name} --policy-arn ${AmazonEC2FullAccess_arn}"
IAMUserSSHKeys_arn="arn:aws:iam::aws:policy/IAMUserSSHKeys"
aws_cmd "iam attach-group-policy --group-name ${group_name} --policy-arn ${IAMUserSSHKeys_arn}"
AmazonS3FullAccess_arn="arn:aws:iam::aws:policy/AmazonS3FullAccess"
aws_cmd "iam attach-group-policy --group-name ${group_name} --policy-arn ${AmazonS3FullAccess_arn}"

# Create bucket for remote state (backer-url)
bucket_name="${4}-${5}-tfstate"
aws_cmd "s3api create-bucket --bucket ${bucket_name}"
echo "BUCKET NAME: ${bucket_name}"
