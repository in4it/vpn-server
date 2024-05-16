#!/bin/bash

AMI_ID=$1
AWS_SUBNET=$2
AWS_SG=$3

#--block-device-mappings "DeviceName=/dev/sda1,Ebs={DeleteOnTermination=true,VolumeSize=30,VolumeType=gp3,Encrypted=false}" \
aws ec2 run-instances \
    --image-id $AMI_ID \
    --instance-type t3.micro \
    --subnet-id $AWS_SUBNET \
    --key-name in4it \
    --ebs-optimized \
    --associate-public-ip-address \
    --security-group-ids $AWS_SG \
    --region us-east-1
