FROM node:16

RUN npm i --location=global aws-cdk

WORKDIR /app

COPY aws-cdk .