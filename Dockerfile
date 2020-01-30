FROM scratch

COPY proxy /proxy
# export CGO_ENABLED=0
EXPOSE 1883
ENTRYPOINT ["/proxy"]