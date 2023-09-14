FROM 		golang:1.20 as build

WORKDIR		/gitlab-aws-credential-helper
ADD		. /gitlab-aws-credential-helper
RUN		CGO_ENABLED=0 GOOS=linux go build  -ldflags '-extldflags "-static"' .

FROM 		alpine:3
COPY --from=build		/gitlab-aws-credential-helper/gitlab-aws-credential-helper  /usr/local/bin
ENTRYPOINT 	["/usr/local/bin/gitlab-aws-credential-helper"]
