# migration-m2-gama

## Install serverless framework
You need to install serverless framewor in your environment
https://www.serverless.com/framework

## Clone repository in your go/src folder

Copy .Makefile content and create a Makefile (this file is similar to .env files) 

On that Makefile you need to change the aws profile for your aws profile created with serverless framework

```
sls deploy --stage=$(stage) --region us-east-1 --aws-profile gaia
```

## How to deploy
```
make deploy stage=stg
```

## Hosts
STG: https://soj713ja6l.execute-api.us-east-1.amazonaws.com/stg

## Endpoints already created
[POST] - {{host}}/users
```
{
  "force": true,
  "users": [
    "maria.valencia@gaiadesign.com.mx",
    "eggcontinued@chewydonut.com",
    "test@reynolds.com",
    "zahit.rios@gaiadesign.com.mx",
    "zahitrios@gmail.com"
  ]
}
```

# Helpful information

## How to create a serverless demo proyect
https://www.serverless.com/framework/docs/tutorial

## How to create your aws profile
https://www.serverless.com/framework/docs/providers/aws/guide/credentials#creating-aws-access-keys

## Proyect examples of go and serverless framework
https://github.com/serverless/examples/tree/master