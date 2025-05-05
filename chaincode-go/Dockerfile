FROM hyperledger/fabric-ccenv:2.4 as builder

USER root

WORKDIR /chaincode
COPY . .

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct

RUN go mod tidy
RUN go build -o energyTrading

FROM hyperledger/fabric-baseos:2.4

WORKDIR /chaincode
COPY --from=builder /chaincode/energyTrading .

CMD ["./energyTrading"]
