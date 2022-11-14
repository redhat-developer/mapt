packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.1"
      source = "github.com/hashicorp/amazon"
    }
  }
}

variable region                   {  default = "us-east-1"}
variable localize                 {  default = "english"}


variable source-admin-pass        {  default = "SuperS3cr3t!!!!"}
variable username                 {  default = "ec2-user"}
variable password                 {  default = "auN=vC%&27ITSj1<"}
variable authorized-keys          {  default = "Required to be fullfilled during userdata exec"}

variable source-ami-name-default  {  default = "Windows_Server-2019-English-Full-HyperV*"}
variable source-ami-name          {
  type = map(string)
  default = {
    "english" = "Windows_Server-2019-English-Full-HyperV*"
    "spanish" = "Windows_Server-2019-Spanish-Full-Base*"
  }
}
variable source-admin-user-default  {  default = "Administrator"}
variable source-admin-user {
  type = map(string)
  default = {
    "english" = "Administrator"
    "spanish" = "Administrador"
  }
}
variable ud-winrm-script-default  {  default = "./ud_winrm_English.ps1"}
variable ud-winrm-script {
  type = map(string)
  default = {
    "english" = "./ud_winrm_English.ps1"
    "spanish" = "./ud_winrm_Spanish.ps1"
  }
}
variable target-ami-name-default  {  default = "Windows_Server-2019-English-Full"}
variable target-ami-name {
  type = map(string)
  default = {
    "english" = "Windows_Server-2019-English-Full"
    "spanish" = "Windows_Server-2019-Spanish-Full"
  }
}

# In case thi variable is not empyt the image generated will included
# openshift local
variable crc-distributable-url  {  default = "" }
variable default-aws-region     {  default= "us-east-1" }

locals {
  if-install-crc              = var.crc-distributable-url != "" ? "amazon-ebs.this" : "none"
  crc-distributable-name      = "crc-windows-installer.zip"
  crc-msi                     = "crc-windows-amd64.msi"

  target-ami-name             = join("-", 
                                  [
                                    lookup(var.target-ami-name, var.localize, var.target-ami-name-default),
                                    var.crc-distributable-url != "" ? "OCPL" : "HyperV",
                                    "RHQE"])

  builder-hyperv-types        = ["t2.small", 
                                  "t2.medium", 
                                  "t3.small", 
                                  "t3.medium"]

  # Openshift local requires at least 9G of RAM to run installation
  builder-ocpl-types          = ["m5zn.xlarge", 
                                  "m5zn.2xlarge",
                                  "m5n.xlarge",
                                  "m5n.2xlarge"]

  #temporary bucket to move assets
  bucket-name                 = "qenvs-packer-${md5(timestamp())}"
  s3-crc-distributable-url    = "https://${local.bucket-name}.s3.${var.default-aws-region}.amazonaws.com/${local.crc-distributable-name}"
}

source "amazon-ebs" "this" {
  ami_name              = local.target-ami-name
  communicator          = "winrm"
  spot_instance_types   =  var.crc-distributable-url != "" ? local.builder-ocpl-types : local.builder-hyperv-types


  # Use spot instance for building process
	spot_price            = "auto"
  region                = var.region

  source_ami_filter {
    filters = {
      name      = lookup(var.source-ami-name, var.localize, var.source-ami-name-default)
    }
    most_recent = true
    owners      = ["amazon"]
  }

  winrm_username        = lookup(var.source-admin-user, var.localize, var.source-admin-user-default) 
  winrm_password        = var.source-admin-pass

  # Recommended property https://developer.hashicorp.com/packer/plugins/builders/amazon/ebs#user_data
  user_data_file        = lookup(var.ud-winrm-script, var.localize, var.ud-winrm-script-default) 
}

build {
  name    = "ol-win"

  sources = ["source.amazon-ebs.this"]

  provisioner powershell {

    elevated_user = lookup(var.source-admin-user, var.localize, var.source-admin-user-default) 
    elevated_password = var.source-admin-pass

    environment_vars = [
      "USERNAME=${var.username}",
      "PASSWORD=${var.password}",
      "AUTHORIZEDKEY=${var.authorized-keys}"]
    script           = "./setup.ps1"
  }

  # Move assets through S3 to avoid winrm slow upload with provisioner file
  # https://github.com/hashicorp/packer/issues/2648
  # Notice this requires aws cli on host node where packer engine is running
  # encourage to use the container image with self contained tools
  provisioner "shell-local" {
    inline  = ["mkdir -p /tmp/tmpdata",
              "wget -q ${var.crc-distributable-url} -O /tmp/tmpdata/${local.crc-distributable-name}",
              "aws s3api create-bucket --bucket ${local.bucket-name} --region ${var.default-aws-region}",
              "aws s3api put-object --bucket ${local.bucket-name} --key ${local.crc-distributable-name} --body /tmp/tmpdata/${local.crc-distributable-name}",
              "aws s3api put-object-acl --bucket ${local.bucket-name} --key ${local.crc-distributable-name} --acl public-read"]

    only    = [local.if-install-crc]
  }

  provisioner powershell {
    inline  = [
      "curl.exe -L ${local.s3-crc-distributable-url} -o C:/Windows/Temp/${local.crc-distributable-name}",
      "Expand-Archive -LiteralPath C:/Windows/Temp/${local.crc-distributable-name} -DestinationPath C:/Windows/Temp -Force",
      "Start-Process C:/Windows/System32/msiexec.exe -ArgumentList '/qb /i C:\\Windows\\Temp\\${local.crc-msi} /norestart' -wait"
    ]

    only    = [local.if-install-crc]
  }

//   $Password = ConvertTo-SecureString "{{.Password}}" -AsPlainText -Force
// $UserAccount = Get-LocalUser -Name "{{.Username}}"
// $UserAccount | Set-LocalUser -Password $Password
// # Also need to set new password on autologin
// $RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
// Set-ItemProperty $RegistryPath 'DefaultPassword' -Value "{{.Password}}" -type String

// # Set the authorized keys according to the private key
// # Due to acl on the files Adminstrator can not write authorized_keys
// # So we need to invoke it as the default user
// $UserName = "{{.Username}}"
// $Cred = New-Object System.Management.Automation.PSCredential ($UserName, $Password)
// Invoke-Command -ScriptBlock { New-Item -Path C:\Users\{{.Username}}\.ssh -Name "authorized_keys" -ItemType "file" -Value "{{.AuthorizedKey}}" -Force } -Credential $Cred -computername localhost

  provisioner powershell {
    inline = [
      # Re-initialise the AWS instance on startup
      # https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/ec2-windows-user-data.html#user-data-scripts-subsequent
      "C:/ProgramData/Amazon/EC2-Windows/Launch/Scripts/InitializeInstance.ps1 -Schedule",
      # Remove system specific information from this image
      # "C:/ProgramData/Amazon/EC2-Windows/Launch/Scripts/SysprepInstance.ps1 -NoShutdown"
    ]
  }

  post-processor manifest {
        output = "manifest.json"
        strip_path = true      
  }

  // TODO postscript to disable winrm
}
