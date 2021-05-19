# GO-AWSDDNS

A quick Dynamic DNS utility that will update a Route53 A record with current IP address.

## Warning
This is very much alpha. I've not done much for errors or edge cases. Use at your own risk.

## Usage
```
Usage of go-awsddns:
  -domain string
        domain name to update ending with . ie test.example.com. 
  -zoneid string
        ZoneID for hosted zone in format /hostedzone/ZONEID
```

## AWS Credentials
The CLI is configured to use default credential chain for AWS. To specify alternate credentials use Environment Variables or the `AWS_PROFILE` environment variable. IE `env AWS_PROFILE=dev awsddns --zoneid example --domain example`
For more information see [here](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials) 
