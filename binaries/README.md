# Prisma binaries

## How to build Prisma CLI binaries

### Setup

Install [zeit/pkg](https://github.com/zeit/pkg):

```shell script
npm i -g pkg
```

### Build the binary and upload to S3

```shell script
mkdir -p build
cd build
npm init --yes
npm i @prisma/cli@2.0.0-alpha.1265

mkdir -p binaries

pkg node_modules/@prisma/cli --out-path binaries/

cd binaries

mv cli-macos "prisma-cli-darwin"
mv cli-linux "prisma-cli-linux"
mv cli-win.exe "prisma-cli-windows.exe"

gzip "prisma-cli-darwin"
gzip "prisma-cli-linux"
gzip "prisma-cli-windows.exe"

aws s3 cp "prisma-cli-$version-darwin.gz" s3://prisma-photongo --acl public-read
aws s3 cp "prisma-cli-$version-linux.gz" s3://prisma-photongo --acl public-read
aws s3 cp "prisma-cli-$version-windows.exe.gz" s3://prisma-photongo --acl public-read
```
