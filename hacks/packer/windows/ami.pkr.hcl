packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.1"
      source = "github.com/hashicorp/amazon"
    }
  }
}

variable region             {  default = "us-east-1"}
variable target-ami-name    {  default = "Windows_Server-2019-English-Full-HyperV-RHQE"}
variable source-ami-name    {  default = "Windows_Server-2019-English-Full-HyperV*"}
variable source-admin-user  {  default = "Administrator"}
variable source-admin-pass  {  default = "SuperS3cr3t!!!!"}
variable ud-winrm-script    {  default = "./ud_winrm_English.ps1"}
variable username           {  default = "ec2-user"}
variable password           {  default = "auN=vC%&27ITSj1<"}
variable authorized-keys    {  default = "Required to be fullfilled during userdata exec"}

source "amazon-ebs" "this" {
  ami_name              = var.target-ami-name
  communicator          = "winrm"
  // instance_type = "t2.micro"
  spot_instance_types   =  ["t2.small", 
    "t2.medium", 
    "t3.small", 
    "t3.medium"]
	spot_price            = "auto"
  region                = var.region

  source_ami_filter {
    filters = {
      name      = var.source-ami-name
    }
    most_recent = true
    owners      = ["amazon"]
  }

  winrm_username        = var.source-admin-user
  winrm_password        = var.source-admin-pass

  user_data_file        = var.ud-winrm-script
}

build {
  name    = "customize"
  sources = ["source.amazon-ebs.this"]

  provisioner powershell {

    elevated_user = var.source-admin-user
    elevated_password = var.source-admin-pass

    environment_vars = [
      "USERNAME=${var.username}",
      "PASSWORD=${var.password}",
      "AUTHORIZEDKEY=${var.authorized-keys}"]
    script           = "./setup.ps1"
  }

  // TODO check function for this
  // provisioner "powershell" {
  //   inline = [
  //     # Re-initialise the AWS instance on startup
  //     "C:/ProgramData/Amazon/EC2-Windows/Launch/Scripts/InitializeInstance.ps1 -Schedule",
  //     # Remove system specific information from this image
  //     "C:/ProgramData/Amazon/EC2-Windows/Launch/Scripts/SysprepInstance.ps1 -NoShutdown"
  //   ]
  // }
  // TODO postscript to disable winrm
}
