FROM alpine:3 as ca
RUN apk add --no-cache ca-certificates

FROM 		golang:1.20 as build

WORKDIR		/gitlab-aws-credential-helper
ADD		. /gitlab-aws-credential-helper
RUN		CGO_ENABLED=0 GOOS=linux go build  -ldflags '-extldflags "-static"' .

FROM 		scratch
ENV PATH=/
COPY --from=ca /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=build		/gitlab-aws-credential-helper/gitlab-aws-credential-helper  /
ENTRYPOINT 	["/gitlab-aws-credential-helper"]
CMD ["dotenv"]
