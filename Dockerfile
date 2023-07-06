FROM docker:20.10.21

RUN apk add bash curl
RUN apk add nodejs npm

RUN npm i -g aws-cdk@2.12.0 && cdk --version
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install


WORKDIR /lamassu-aws-cdk
COPY aws-cdk/package.json package.json
COPY aws-cdk/package-lock.json package-lock.json

RUN npm install
# Without this npm dependency, the cdk deploy fails with: throw new Error(`Failed to bundle asset ${this.node.path}, bundle output is located at ${bundleErrorDir}: ${err}`);
RUN npm i -g esbuild 

COPY aws-cdk .

RUN npm run build
COPY docker-entrypoint.sh /lamassu-aws-cdk/docker-entrypoint.sh

ENTRYPOINT ["/bin/bash", "docker-entrypoint.sh"]